package pwsafe

import (
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/twofish"
)

// Decrypt Decrypts the data in the reader using the given password and populates the information into the db
func (db *V3) Decrypt(reader io.Reader, passwd string) (int, error) {
	// read the entire encrypted db into memory
	var rawDB []byte
	var bytesRead int
	for {
		block := make([]byte, 256)
		readLoop, err := reader.Read(block)
		bytesRead += readLoop
		if err != nil {
			return bytesRead, err
		}
		rawDB = append(rawDB, block[:readLoop]...)
		if readLoop != 256 {
			break
		}
	}

	if bytesRead < 200 {
		return bytesRead, errors.New("DB file is smaller than minimum size")
	}

	// The TAG is 4 ascii characters, should be "PWS3"
	if string(rawDB[:4]) != "PWS3" {
		return bytesRead, errors.New("File is not a valid Password Safe v3 file")
	}
	pos := 4 // used to track the current position in the byte array representing the db.

	// Read the Salt
	copy(db.Salt[:], rawDB[pos:pos+32])
	pos += 32

	// Read iter
	db.Iter = uint32(byteToInt(rawDB[pos : pos+4]))
	pos += 4

	// Verify the password
	db.calculateStretchKey(passwd)
	var keyHash [sha256.Size]byte
	copy(keyHash[:], rawDB[pos:pos+sha256.Size])
	pos += sha256.Size
	if keyHash != sha256.Sum256(db.StretchedKey[:]) {
		return bytesRead, errors.New("Invalid Password")
	}

	//extract the encryption and hmac keys
	db.extractKeys(rawDB[pos : pos+64])
	pos += 64

	copy(db.CBCIV[:], rawDB[pos:pos+16])
	pos += 16

	// All following fields are encrypted with twofish in CBC mode until the EOF, find the EOF
	var encryptedDB []byte
	var encryptedSize int
	for {
		if pos+twofish.BlockSize > bytesRead {
			return bytesRead, errors.New("Invalid DB, no EOF found")
		}
		blockBytes := rawDB[pos : pos+twofish.BlockSize]
		pos += twofish.BlockSize

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
	expectedHMAC := rawDB[pos : pos+32]
	if len(rawDB) != pos+32 {
		return bytesRead, errors.New("Error unknown data after expected EOF")
	}

	//UnMarshal the decrypted DB, first the header
	header, hdrSize, headerHMACData, err := UnmarshalHeader(decryptedDB)
	if err != nil {
		return bytesRead, errors.New("Error parsing the unencrypted header - " + err.Error())
	}
	db.Header = header

	_, recordHMACData, err := db.unmarshalRecords(decryptedDB[hdrSize:])
	if err != nil {
		return bytesRead, errors.New("Error parsing the unencrypted records - " + err.Error())
	}
	hmacData := append(headerHMACData, recordHMACData...)

	// Verify HMAC - The HMAC is only calculated on the header/field values not length/type
	db.calculateHMAC(hmacData)
	if !hmac.Equal(db.HMAC[:], expectedHMAC) {
		return bytesRead, errors.New("Error Calculated HMAC does not match read HMAC")
	}

	return bytesRead, nil
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
		fieldLength := byteToInt(records[fieldStart : fieldStart+4])
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
