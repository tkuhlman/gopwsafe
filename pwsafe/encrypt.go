package pwsafe

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	pseudoRand "math/rand"
	"reflect"
	"time"

	"github.com/pborman/uuid"

	"golang.org/x/crypto/twofish"
)

// Encrypt Encrypt the data in the db building it up in memory then writing to the writer, returns bytesWritten, error
func (db *V3) Encrypt(dbBuf io.Writer) error {
	//update the LastSave time in the DB
	db.Header.LastSave = time.Now()
	db.Header.Version = [2]byte{0x10, 0x03} // DB Format version 0x0310

	// Set unencrypted DB headers
	if err := binary.Write(dbBuf, binary.LittleEndian, []byte("PWS3")); err != nil {
		return err
	}

	// Add salt and iter neither of which can change without knowing the password as the stretchedkey will need recalculating.
	// use db.SetPassword() to change the password
	if err := binary.Write(dbBuf, binary.LittleEndian, db.Salt); err != nil {
		return err
	}
	if err := binary.Write(dbBuf, binary.LittleEndian, db.Iter); err != nil {
		return err
	}

	// Add the stretchedKey Hash and refresh the encryption keys adding them encrypted
	stretchedSHA := sha256.Sum256(db.StretchedKey[:])
	if err := binary.Write(dbBuf, binary.LittleEndian, stretchedSHA); err != nil {
		return err
	}
	if err := db.refreshEncryptedKeys(dbBuf); err != nil {
		return err
	}

	// calculate and add cbc initial value
	_, err := rand.Read(db.CBCIV[:])
	if err != nil {
		return err
	}
	if err := binary.Write(dbBuf, binary.LittleEndian, db.CBCIV); err != nil {
		return err
	}

	// marshal the core db values

	/* TODO consider switching to a bytes buffer
	unencryptedBuf := &bytes.Buffer{}

	// TODO MarshalBinary will not work because I need to return to calculate the HMAC bytes as well, see the function to caculate the HMAC and continue from there.
	header, err := db.Header.MarshalBinary(hmacBuf)
	if err!= nil {
		return fmt.Errorf("error marshalling header: %w", err)
	}
	if _, err := unencryptedBuf.Write(header); err != nil {
		return err
	}
	*/
	var unencryptedBytes []byte
	// Note the version field needs to be first and is required
	// headerFields := structs.Fields(db.Header)
	//todo it is a bad assumption that version is the last item in the slice, fix so version is first
	//ordered := structs.Fields(db)
	//headerFields := append(ordered[:len(ordered)-2], ordered[len(ordered)-1])

	headerBytes, headerValues := marshalRecord(reflect.ValueOf(db.Header))
	unencryptedBytes = append(unencryptedBytes, headerBytes...)

	recordBytes, recordValues := db.marshalRecords()
	unencryptedBytes = append(unencryptedBytes, recordBytes...)

	// encrypt and write the dbBlocks
	dbTwoFish, _ := twofish.NewCipher(db.EncryptionKey[:])
	cbcTwoFish := cipher.NewCBCEncrypter(dbTwoFish, db.CBCIV[:])
	for i := 0; i < len(unencryptedBytes); i += twofish.BlockSize {
		block := unencryptedBytes[i : i+twofish.BlockSize]
		encrypted := make([]byte, twofish.BlockSize)
		cbcTwoFish.CryptBlocks(encrypted, block)
		if err := binary.Write(dbBuf, binary.LittleEndian, encrypted); err != nil {
			return err
		}
	}

	// Add the EOF and HMAC
	if err := binary.Write(dbBuf, binary.LittleEndian, []byte("PWS3-EOFPWS3-EOF")); err != nil {
		return err
	}
	hmacBytes := append(headerValues, recordValues...)
	db.calculateHMAC(hmacBytes)
	if err := binary.Write(dbBuf, binary.LittleEndian, db.HMAC); err != nil {
		return err
	}

	return nil
}

// For the given field return the []byte representation of its data
func getFieldBytes(field reflect.Value) (fbytes []byte) {
	switch field.Kind() {
	// switch field.Kind()
	// case reflect.
	case reflect.String:
		fstring := field.String()
		fbytes = []byte(fstring)
	case reflect.Struct: //time.Time shows as kind struct
		if field.Type() == reflect.TypeOf(time.Time{}) {
			fbytes = intToBytes(int(field.Interface().(time.Time).Unix()))
		}
	case reflect.Array:
		switch field.Len() {
		case 2:
			farray := field.Interface().([2]byte)
			fbytes = farray[:]
		case 4:
			farray := field.Interface().([4]byte)
			fbytes = farray[:]
		case 16:
			farray := field.Interface().([16]byte)
			fbytes = farray[:]
		}
	default:
		if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.Uint8 {
			fbytes = field.Bytes()
		}
	}
	return fbytes
}

// marshalHeader return the binary format for the record as specified in the spec and the header values used for hmac calculations
// This function is used both to Marshal the header and individual records in the DB
func marshalRecord(s reflect.Value) (record []byte, totalDataBytes []byte) {
	// TODO the UUID, password and title fields are mandatory, make sure the are defined before a record can be added to the DB.
	typ := s.Type()
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		typeField := typ.Field(i)
		fieldTypeStr := typeField.Tag.Get("field")

		if fieldTypeStr == "" || isZero(field) {
			continue
		} else {
			fieldType, err := hex.DecodeString(fieldTypeStr)
			if err != nil {
				panic(fmt.Sprintf("Invalid field type in struct tag for %s\n\t%v", typeField.Name, err))
			}
			dataBytes := getFieldBytes(field)
			totalDataBytes = append(totalDataBytes, dataBytes...)

			// Each record is the length, type and data
			record = append(record, intToBytes(len(dataBytes))...)
			record = append(record, fieldType[0])

			// Add in the data
			record = append(record, dataBytes...)

			// if total written bytes doesn't match twofish.BlockSize fill remaining bytes with pseudo random values
			usedBlockSpace := (len(dataBytes) + 5) % twofish.BlockSize
			if usedBlockSpace != 0 {
				record = append(record, pseudoRandmonBytes(twofish.BlockSize-usedBlockSpace)...)
			}
		}
	}

	//finish with the end of record
	record = append(record, []byte{0, 0, 0, 0}...)
	record = append(record, '\xFF')
	record = append(record, pseudoRandmonBytes(twofish.BlockSize-5)...)

	return record, totalDataBytes
}

// isZero checks if a value is zero-valued
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		// Special handling for time.Time
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	default:
		// Compare with zero value of the type
		zero := reflect.Zero(v.Type())
		return v.Interface() == zero.Interface()
	}
}

// marshalRecords return the binary format for the Records as specified in the spec and the record values used for hmac calculations
func (db *V3) marshalRecords() (records []byte, dataBytes []byte) {
	for _, record := range db.Records {
		recordVal := reflect.ValueOf(record)
		// if uuid is not set calculate
		//todo I should assume the UUID is set. I do for new dbs but don't check on reading from disk, I
		// should check it is unique also when opening more than one in the gui
		if isZero(recordVal.FieldByName("UUID")) {
			db.Header.UUID = [16]byte(uuid.NewRandom().Array())
		}

		// for each record UUID, Title and Password fields are mandatory all others are optional
		if isZero(recordVal.FieldByName("Title")) || isZero(recordVal.FieldByName("Password")) {
			//todo how should I handle this?
			fmt.Println("Error: Title or Password is not set, invalid record")
			continue
		}

		// finally call marshalRecord for this record
		rBytes, hmacBytes := marshalRecord(recordVal)
		records = append(records, rBytes...)
		dataBytes = append(dataBytes, hmacBytes...)
	}

	return records, dataBytes
}

// Generate size bytes of pseudo random data
func pseudoRandmonBytes(size int) (r []byte) {
	for i := 0; i < size; i += 8 {
		bytesRand := make([]byte, 16)
		binary.PutVarint(bytesRand, pseudoRand.Int63())
		r = append(r, bytesRand...)
	}
	return r[:size]
}

// re-calculate and add to the db new encryption key and hmac key then encrypt with and return the encrypted bytes
func (db *V3) refreshEncryptedKeys(buf io.Writer) error {
	_, err := rand.Read(db.EncryptionKey[:])
	if err != nil {
		return err
	}
	_, err = rand.Read(db.HMACKey[:])
	if err != nil {
		return err
	}
	keyTwoFish, err := twofish.NewCipher(db.StretchedKey[:])
	if err != nil {
		return err
	}
	for _, block := range [][]byte{db.EncryptionKey[:16], db.EncryptionKey[16:], db.HMACKey[:16], db.HMACKey[16:]} {
		encrypted := make([]byte, 16)
		keyTwoFish.Encrypt(encrypted, block)
		if err := binary.Write(buf, binary.LittleEndian, encrypted); err != nil {
			return err
		}
	}
	return nil
}

// TODO get rid of this just use binary.Write
// intToBytes Converts an int to byte array
func intToBytes(num int) []byte {
	intBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(intBytes, uint32(num))
	return intBytes
}
