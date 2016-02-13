package pwsafe

import (
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"io"

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

// Parse the header of the decrypted DB returning the size of the Header, a byte array for calculating hmac and any error or nil
// beginning with the Version type field, and terminated by the 'END' type field. The version number
// and END fields are mandatory
func (db *V3) parseHeader(decryptedDB []byte) (int, []byte, error) {
	fieldStart := 0
	var hmacData []byte
	for {
		if fieldStart > len(decryptedDB) {
			return 0, hmacData, errors.New("No END field found in DB header")
		}
		fieldLength := byteToInt(decryptedDB[fieldStart : fieldStart+4])
		btype := byteToInt(decryptedDB[fieldStart+4 : fieldStart+5])

		data := decryptedDB[fieldStart+5 : fieldStart+fieldLength+5]
		hmacData = append(hmacData, data...)
		fieldStart += fieldLength + 5
		//The next field must start on a block boundary
		blockmod := fieldStart % twofish.BlockSize
		if blockmod != 0 {
			fieldStart += twofish.BlockSize - blockmod
		}
		switch btype {
		case 0x00: //version
			db.Version = string(data)
		case 0x01: //uuuid
			db.UUID = data
		case 0x02: //preferences
			continue
		case 0x03: //tree
			continue
		case 0x04: //timestamp
			continue
		case 0x05: //who last save
			continue
		case 0x06: //last save timestamp
			continue
		case 0x07: //last save user
			continue
		case 0x08: //last save host
			continue
		case 0x09: //DB name
			db.Name = string(data)
		case 0x0a: //description
			db.Description = string(data)
		case 0x0b: //filters
			continue
		case 0x0f: //recently used
			continue
		case 0x10: //password policy
			continue
		case 0x11: //Empty Groups
			continue
		case 0xff: //end
			return fieldStart, hmacData, nil
		default:
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
	fieldStart := 0
	var record Record
	var rdata []byte
	for {
		fieldLength := byteToInt(records[fieldStart : fieldStart+4])
		btype := byteToInt(records[fieldStart+4 : fieldStart+5])
		data := records[fieldStart+5 : fieldStart+fieldLength+5]
		rdata = append(rdata, data...)
		fieldStart += fieldLength + 5
		//The next field must start on a block boundary
		blockmod := fieldStart % twofish.BlockSize
		if blockmod != 0 {
			fieldStart += twofish.BlockSize - blockmod
		}
		switch btype {
		case 0x01:
			record.UUID = data
		case 0x02:
			record.Group = string(data)
		case 0x03:
			record.Title = string(data)
		case 0x04:
			record.Username = string(data)
		case 0x05:
			record.Notes = string(data)
		case 0x06:
			record.Password = string(data)
		case 0x07:
			continue
		case 0x08:
			continue
		case 0x09:
			continue
		case 0x0a: // password expiry time
			continue
		case 0x0c:
			continue
		case 0x0d:
			record.URL = string(data)
		case 0x0e: //autotype
			continue
		case 0x0f: //password history
			continue
		case 0x10: //password policy
			continue
		case 0x11: //password expiry interval
			continue
		case 0x12: //run command
			continue
		case 0x13: //double click action
			continue
		case 0x14: //email
			continue
		case 0x15: //protected entry
			continue
		case 0x16: //own symbol
			continue
		case 0x17: //shift double click action
			continue
		case 0x18: //password policy name
			continue
		case 0xff: //end
			db.Records[record.Title] = record
			return fieldStart, rdata, nil
		default:
			return fieldStart, rdata, errors.New("Encountered unknown Record Field type - " + string(btype))
		}
	}
}
