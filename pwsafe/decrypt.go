package pwsafe

import (
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

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

// mapByFieldTag Return map[byte]reflect.Value for a struct where byte is the "field" struct tag converted to a byte
// if field struct tag doesn't exist skip that field
func mapByFieldTag(s interface{}) map[byte]reflect.Value {
	fieldMap := make(map[byte]reflect.Value)
	val := reflect.ValueOf(s).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)
		fieldTypeStr := typeField.Tag.Get("field")
		if fieldTypeStr == "" {
			continue
		}
		fieldType, err := hex.DecodeString(fieldTypeStr)
		if err != nil {
			panic(fmt.Sprintf("Invalid field type in struct tag for %s\n\t%v", typeField.Name, err))
		}
		if len(fieldType) > 0 {
			fieldMap[fieldType[0]] = field
		}
	}
	return fieldMap
}

// setField Set the value of the Field with the proper conversion for its type
func setField(field reflect.Value, data []byte) {
	if !field.CanSet() {
		// This should ideally not happen if mapByFieldTag correctly returns settable fields.
		// Or if the struct passed to mapByFieldTag is a pointer.
		panic(fmt.Sprintf("Cannot set field %s", field.Type().Name()))
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(string(data))
	case reflect.Struct: //time.Time shows as kind struct
		if field.Type() == reflect.TypeOf(time.Time{}) {
			field.Set(reflect.ValueOf(time.Unix(int64(byteToInt(data)), 0)))
		}
	case reflect.Array, reflect.Slice:
		switch len(data) {
		case 2:
			var farray [2]byte
			copy(farray[:], data)
			field.Set(reflect.ValueOf(farray))
		case 4:
			var farray [4]byte
			copy(farray[:], data)
			field.Set(reflect.ValueOf(farray))
		case 16:
			var farray [16]byte
			copy(farray[:], data)
			field.Set(reflect.ValueOf(farray))
		}
	case reflect.Uint8: // For byte fields
		if len(data) == 1 {
			field.SetUint(uint64(data[0]))
		}
	default:
		// For []byte or other types, we might need more specific handling if they are not covered above.
		// Assuming []byte for now if it matches the type.
		if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.Uint8 {
			field.SetBytes(data)
		} else if field.Type() == reflect.TypeOf([]byte{}) { // Explicit check for []byte type
			field.SetBytes(data)
		} else {
			// If we reach here, it's an unhandled type or a type that cannot be directly set from []byte
			// without further conversion. Panic for now to indicate an unhandled case.
			panic(fmt.Sprintf("Unhandled field type for setting: %s (Kind: %s)", field.Type().Name(), field.Kind()))
		}
	}
}

// UnMarshal the records returning records length, a byte array of data for hmac calculations and error or nil
// The EOF string records end with is "PWS3-EOFPWS3-EOF"
func (db *V3) unmarshalRecords(records []byte) (int, []byte, error) {
	recordStart := 0
	var hmacData []byte
	db.Records = make(map[string]Record)
	for recordStart < len(records) {
		record := &Record{}
		recordFieldMap := mapByFieldTag(record)
		recordLength, recordData, err := unmarshalRecord(records[recordStart:], recordFieldMap)
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
func unmarshalRecord(records []byte, recordFieldMap map[byte]reflect.Value) (int, []byte, error) {
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

		field, prs := recordFieldMap[btype]
		if prs {
			setField(field, data)
		} else if btype == 0xff { //end
			return fieldStart, rdata, nil
		} else {
			return fieldStart, rdata, fmt.Errorf("Encountered unknown Record Field type - %v", btype)
		}
	}
}
