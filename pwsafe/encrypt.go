package pwsafe

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	pseudoRand "math/rand"
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

	var unencryptedBytes []byte
	// Note the version field needs to be first and is required

	headerBytes, headerValues := db.Header.marshal()
	unencryptedBytes = append(unencryptedBytes, headerBytes...)

	recordBytes, recordValues, err := db.marshalRecords()
	if err != nil {
		return err
	}
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

// marshalRecords return the binary format for the Records as specified in the spec and the record values used for hmac calculations
func (db *V3) marshalRecords() (records []byte, dataBytes []byte, err error) {
	for _, record := range db.Records {
		// if uuid is not set calculate
		//todo I should assume the UUID is set. I do for new dbs but don't check on reading from disk, I
		// should check it is unique also when opening more than one in the gui
		if record.UUID == [16]byte{} {
			db.Header.UUID = [16]byte(uuid.NewRandom().Array())
		}

		// for each record UUID, Title and Password fields are mandatory all others are optional
		if record.Title == "" || record.Password == "" {
			return nil, nil, fmt.Errorf("Title or Password is not set, invalid record, title %s", record.Title)
		}

		// finally call marshalRecord for this record
		rBytes, hmacBytes := record.marshal()
		records = append(records, rBytes...)
		dataBytes = append(dataBytes, hmacBytes...)
	}

	return records, dataBytes, nil
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
