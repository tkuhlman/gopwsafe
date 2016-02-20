package pwsafe

import (
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/fatih/structs"
	"golang.org/x/crypto/twofish"
)

//Decrypt Decrypts the data in the reader using the given password and populates the information into the db
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
	if keyHash != sha256.Sum256(db.stretchedKey[:]) {
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

	block, err := twofish.NewCipher(db.encryptionKey[:])
	decrypter := cipher.NewCBCDecrypter(block, db.CBCIV[:])
	decryptedDB := make([]byte, encryptedSize) // The EOF and HMAC are after the encrypted section
	decrypter.CryptBlocks(decryptedDB, encryptedDB)

	// Verify expected end of data
	expectedHMAC := rawDB[pos : pos+32]
	if len(rawDB) != pos+32 {
		return bytesRead, errors.New("Error unknown data after expected EOF")
	}

	//Parse the decrypted DB, first the header
	hdrSize, headerHMACData, err := db.parseHeader(decryptedDB)
	if err != nil {
		return bytesRead, errors.New("Error parsing the unencrypted header - " + err.Error())
	}

	_, recordHMACData, err := db.parseRecords(decryptedDB[hdrSize:])
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
	c, _ := twofish.NewCipher(db.stretchedKey[:])
	k1 := make([]byte, 16)
	c.Decrypt(k1, keyData[:16])
	k2 := make([]byte, 16)
	c.Decrypt(k2, keyData[16:32])
	copy(db.encryptionKey[:], append(k1, k2...))

	l1 := make([]byte, 16)
	c.Decrypt(l1, keyData[32:48])
	l2 := make([]byte, 16)
	c.Decrypt(l2, keyData[48:])
	copy(db.HMACKey[:], append(l1, l2...))
}

// mapByFieldTag Return map[byte]*structs.Field for a struct where byte is the "field" struct tag converted to a byte
// if field struct tag doesn't exist skip that field
func mapByFieldTag(s interface{}) map[byte]*structs.Field {
	fieldMap := make(map[byte]*structs.Field)
	for _, field := range structs.Fields(s) {
		fieldTypeStr := field.Tag("field")
		fieldType, err := hex.DecodeString(fieldTypeStr)
		if err != nil {
			panic(fmt.Sprintf("Invalid field type in struct tag for %s\n\t%v", field.Name(), err))
		}
		if len(fieldType) > 0 {
			fieldMap[fieldType[0]] = field
		}
	}
	return fieldMap
}

// Parse the header of the decrypted DB returning the size of the Header, a byte array for calculating hmac and any error or nil
// beginning with the Version type field, and terminated by the 'END' type field. The version number
// and END fields are mandatory
func (db *V3) parseHeader(decryptedDB []byte) (int, []byte, error) {
	fieldStart := 0
	dbFieldMap := mapByFieldTag(db)
	var hmacData []byte
	for {
		if fieldStart > len(decryptedDB) {
			return 0, hmacData, errors.New("No END field found in DB header")
		}
		fieldLength := byteToInt(decryptedDB[fieldStart : fieldStart+4])
		btype := decryptedDB[fieldStart+4 : fieldStart+5][0]

		data := decryptedDB[fieldStart+5 : fieldStart+fieldLength+5]
		hmacData = append(hmacData, data...)
		fieldStart += fieldLength + 5
		//The next field must start on a block boundary
		blockmod := fieldStart % twofish.BlockSize
		if blockmod != 0 {
			fieldStart += twofish.BlockSize - blockmod
		}

		field, prs := dbFieldMap[btype]
		if prs {
			setField(field, data)
		} else if btype == 0xff { //end
			return fieldStart, hmacData, nil
		} else {
			return 0, hmacData, errors.New("Encountered unknown Header Field " + string(btype))
		}

	}
}

// Parse the records returning records length, a byte array of data for hmac calculations and error or nil
// The EOF string records end with is "PWS3-EOFPWS3-EOF"
func (db *V3) parseRecords(records []byte) (int, []byte, error) {
	recordStart := 0
	var hmacData []byte
	db.Records = make(map[string]Record)
	for recordStart < len(records) {
		recordLength, recordData, err := db.parseNextRecord(records[recordStart:])
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

// Parse a single record from the given records []byte, return record size, raw record Data and error/nil
// Individual records stop with an END filed and UUID, Title and Password fields are mandatory all others are optional
func (db *V3) parseNextRecord(records []byte) (int, []byte, error) {
	record := &Record{}
	var rdata []byte
	fieldStart := 0
	recordFieldMap := mapByFieldTag(record)
	for {
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

		field, prs := recordFieldMap[btype]
		if prs {
			setField(field, data)
		} else if btype == 0xff { //end
			db.Records[record.Title] = *record
			return fieldStart, rdata, nil
		} else {
			return fieldStart, rdata, fmt.Errorf("Encountered unknown Record Field type - %v", btype)
		}
	}
}

// setField Set the value of the Field with the proper conversion for its type
func setField(field *structs.Field, data []byte) {
	switch field.Kind().String() {
	case "string":
		err := field.Set(string(data))
		if err != nil {
			panic(err)
		}
		// case uuid.uuid  // this may not need to be specialy dealt with
	case "struct": //time.Time shows as kind struct
		err := field.Set(time.Unix(int64(byteToInt(data)), 0))
		if err != nil {
			panic(err)
		}
	default:
		err := field.Set(data)
		if err != nil {
			panic(err)
		}
	}
}
