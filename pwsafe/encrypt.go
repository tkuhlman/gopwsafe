package pwsafe

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"time"

	"golang.org/x/crypto/twofish"
)

//Encrypt Encrypt the data in the db building it up in memory then writing to the writer, returns bytesWritten, error
func (db *V3) Encrypt(writer io.Writer) (int, error) {
	var dbBytes []byte

	// Set unencrypted DB headers
	dbBytes = append(dbBytes, "PWS3"...)

	//update the LastSave time in the DB
	db.LastSave = time.Now()

	// Add salt and iter neither of which can change without knowing the password as the stretchedkey will need recalculating.
	// use db.SetPassword() to change the password
	dbBytes = append(dbBytes, db.Salt[:]...)
	iter := make([]byte, 4)
	binary.LittleEndian.PutUint32(iter, db.Iter)
	dbBytes = append(dbBytes, iter...)

	// Add the stretchedKey
	stretchedSha := sha256.Sum256(db.stretchedKey[:])
	dbBytes = append(dbBytes, stretchedSha[:]...)

	// re-calculate, encrypt and add encryption key and hmac key
	_, err := rand.Read(db.encryptionKey[:])
	if err != nil {
		return 0, err
	}
	_, err = rand.Read(db.HMACKey[:])
	if err != nil {
		return 0, err
	}
	keyTwoFish, _ := twofish.NewCipher(db.stretchedKey[:])
	for _, block := range [][]byte{db.encryptionKey[:16], db.encryptionKey[16:], db.HMACKey[:16], db.HMACKey[16:]} {
		encrypted := make([]byte, 16)
		keyTwoFish.Encrypt(encrypted, block)
		dbBytes = append(dbBytes, encrypted...)
	}

	// calculate and add cbc initial value
	_, err = rand.Read(db.CBCIV[:])
	if err != nil {
		return 0, err
	}
	dbBytes = append(dbBytes, db.CBCIV[:]...)

	// marshall the core db valudes
	var unencryptedBytes []byte
	headerBytes, headerValues := db.marshallHeader()
	unencryptedBytes = append(unencryptedBytes, headerBytes...)
	recordBytes, recordValues := db.marshallRecords()
	unencryptedBytes = append(unencryptedBytes, recordBytes...)

	// encrypt and write the dbBlocks
	dbTwoFish, _ := twofish.NewCipher(db.encryptionKey[:])
	for i := 0; i < len(unencryptedBytes); i += 16 {
		block := unencryptedBytes[i : i+16]
		encrypted := make([]byte, 16)
		dbTwoFish.Encrypt(encrypted, block)
		dbBytes = append(dbBytes, encrypted...)
	}

	// Add the EOF and HMAC
	dbBytes = append(dbBytes, []byte("PWS3-EOFPWS3-EOF")...)
	hmacBytes := append(headerValues, recordValues...)
	db.calculateHMAC(hmacBytes)
	dbBytes = append(dbBytes, db.HMAC[:]...)

	// Write out the db
	// todo - skip the write until we have an actual valid db implemented
	return 0, nil
	//bytesWritten, err : writer.Write(dbBrytes)
	//return bytesWritten, err
}

// marshallHeader return the binary format for the Header as specified in the spec and the header values used for hmac calculations
func (db *V3) marshallHeader() ([]byte, []byte) {

	//  ** Note the version field needs to be first

	// ideas
	//	st := reflect.TypeOf(db)
	// for i < st.NumField()
	//	field := st.Field(0)
	//	btype := field.Tag.Get("field") //todo this is a hex string make it a byte.
	//if btype == '' ; skip
	// write field length
	// write btype
	// convert (if field.Kind == X; field.ValueOf)and write data - add data to the hmacValues
	// if total written bytes doesn't match twofish.BlockSize fill remaining bytes with pseudo random values

	return []byte("unimplemented"), []byte("unimplemented")
}

// marshallRecords return the binary format for the Records as specified in the spec and the record values used for hmac calculations
func (db *V3) marshallRecords() ([]byte, []byte) {
	return []byte("unimplemented"), []byte("unimplemented")
}
