package pwsafe

import (
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/twofish"
)

// Decrypt Decrypts the data in the reader using the given password and populates the information into the db
func (db *V3) Decrypt(reader io.Reader, passwd string) (int, error) {
	cr := &CountingReader{Reader: reader}

	// The TAG is 4 ascii characters, should be "PWS3"
	tag := make([]byte, 4)
	if _, err := io.ReadFull(cr, tag); err != nil {
		return cr.BytesRead, err
	}
	if string(tag) != "PWS3" {
		return cr.BytesRead, errors.New("File is not a valid Password Safe v3 file")
	}

	// Read the Salt
	if _, err := io.ReadFull(cr, db.Salt[:]); err != nil {
		return cr.BytesRead, err
	}

	// Read iter
	if err := binary.Read(cr, binary.LittleEndian, &db.Iter); err != nil {
		return cr.BytesRead, err
	}

	// Verify the password
	db.calculateStretchKey(passwd)
	var keyHash [sha256.Size]byte
	if _, err := io.ReadFull(cr, keyHash[:]); err != nil {
		return cr.BytesRead, err
	}
	if keyHash != sha256.Sum256(db.StretchedKey[:]) {
		return cr.BytesRead, errors.New("Invalid Password")
	}

	//extract the encryption and hmac keys
	keyData := make([]byte, 64)
	if _, err := io.ReadFull(cr, keyData); err != nil {
		return cr.BytesRead, err
	}
	db.extractKeys(keyData)

	if _, err := io.ReadFull(cr, db.CBCIV[:]); err != nil {
		return cr.BytesRead, err
	}

	// All following fields are encrypted with twofish in CBC mode until the EOF, find the EOF
	var encryptedDB []byte
	var encryptedSize int
	for {
		blockBytes := make([]byte, twofish.BlockSize)
		if _, err := io.ReadFull(cr, blockBytes); err != nil {
			return cr.BytesRead, err
		}

		if string(blockBytes) == "PWS3-EOFPWS3-EOF" {
			break
		} else {
			encryptedSize += twofish.BlockSize
			encryptedDB = append(encryptedDB, blockBytes...)
		}
	}

	block, err := twofish.NewCipher(db.EncryptionKey[:])
	if err != nil {
		return 0, err
	}
	decrypter := cipher.NewCBCDecrypter(block, db.CBCIV[:])
	decryptedDB := make([]byte, encryptedSize) // The EOF and HMAC are after the encrypted section
	decrypter.CryptBlocks(decryptedDB, encryptedDB)

	// Verify expected end of data
	expectedHMAC := make([]byte, 32)
	if _, err := io.ReadFull(cr, expectedHMAC); err != nil {
		return cr.BytesRead, err
	}

	//UnMarshal the decrypted DB, first the header
	header, hdrSize, headerHMACData, err := UnmarshalHeader(decryptedDB)
	if err != nil {
		return cr.BytesRead, errors.New("Error parsing the unencrypted header - " + err.Error())
	}
	db.Header = header

	_, recordHMACData, err := db.unmarshalRecords(decryptedDB[hdrSize:])
	if err != nil {
		return cr.BytesRead, errors.New("Error parsing the unencrypted records - " + err.Error())
	}
	hmacData := append(headerHMACData, recordHMACData...)

	// Verify HMAC - The HMAC is only calculated on the header/field values not length/type
	db.calculateHMAC(hmacData)
	if !hmac.Equal(db.HMAC[:], expectedHMAC) {
		return cr.BytesRead, errors.New("Error Calculated HMAC does not match read HMAC")
	}

	return cr.BytesRead, nil
}

// CountingReader wraps an io.Reader and counts the bytes read
type CountingReader struct {
	io.Reader
	BytesRead int
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.BytesRead += n
	return n, err
}

// Pull encryptionKey and HMAC key from the 64byte keyData
func (db *V3) extractKeys(keyData []byte) {
	c, _ := twofish.NewCipher(db.StretchedKey[:])
	k1 := make([]byte, 16)
	c.Decrypt(k1, keyData[:16])
	k2 := make([]byte, 16)
	c.Decrypt(k2, keyData[16:32])
	copy(db.EncryptionKey[:], append(k1, k2...))

	l1 := make([]byte, 16)
	c.Decrypt(l1, keyData[32:48])
	l2 := make([]byte, 16)
	c.Decrypt(l2, keyData[48:])
	copy(db.HMACKey[:], append(l1, l2...))
}

// fieldSetter interface for types that can set fields by ID
type fieldSetter interface {
	setField(id byte, data []byte) error
}

// UnMarshal the records returning records length, a byte array of data for hmac calculations and error or nil
// The EOF string records end with is "PWS3-EOFPWS3-EOF"
func (db *V3) unmarshalRecords(records []byte) (int, []byte, error) {
	recordStart := 0
	var hmacData []byte
	db.Records = make(map[string]Record)
	for recordStart < len(records) {
		record := &Record{}
		recordLength, recordData, err := unmarshalRecord(records[recordStart:], record)
		db.Records[record.Title] = *record
		if err != nil {
			return recordStart, hmacData, errors.New("Error parsing record - " + err.Error())
		}
		hmacData = append(hmacData, recordData...)
		recordStart += recordLength
	}

	if recordStart > len(records) {
		return recordStart, hmacData, errors.New("Encountered a record with invalid length")
	}
	return recordStart, hmacData, nil
}

// UnMarshal a single record from the given records []byte, writing to fields in recordFieldMap, return record size, raw record Data and error or nil
// Individual records stop with an END field
// This function is used both to UnMarshal the header and individual records in the DB
func unmarshalRecord(records []byte, setter fieldSetter) (int, []byte, error) {
	var rdata []byte
	fieldStart := 0
	for {
		if fieldStart > len(records) {
			return 0, rdata, errors.New("No END field found when UnMarshaling")
		}
		fieldLength := int(binary.LittleEndian.Uint32(records[fieldStart : fieldStart+4]))
		btype := records[fieldStart+4 : fieldStart+5][0]
		data := records[fieldStart+5 : fieldStart+fieldLength+5]
		rdata = append(rdata, data...)
		fieldStart += fieldLength + 5
		//The next field must start on a block boundary
		blockmod := fieldStart % twofish.BlockSize
		if blockmod != 0 {
			fieldStart += twofish.BlockSize - blockmod
		}

		if btype == recordEndOfEntry { // Using RecordEndOfEntry as generic end marker, assuming it's same for header
			return fieldStart, rdata, nil
		}

		if err := setter.setField(btype, data); err != nil {
			return fieldStart, rdata, err
		}
	}
}
