package pwsafe

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"time"

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

		// for each record UUID, Title and Password fields are mandatory all others are optional
		if record.Title == "" || record.Password == "" {
			return nil, nil, fmt.Errorf("title or password is not set, invalid record, title %s", record.Title)
		}

		// finally call marshalRecord for this record
		rBytes, hmacBytes, err := record.marshal()
		if err != nil {
			return nil, nil, err
		}
		records = append(records, rBytes...)
		dataBytes = append(dataBytes, hmacBytes...)
	}

	return records, dataBytes, nil
}

// Generate size bytes of pseudo random data
// Generate size bytes of pseudo random data
func pseudoRandomBytes(size int) (r []byte) {
	r = make([]byte, size)
	_, err := rand.Read(r)
	if err != nil {
		// Fallback to zero padding if rand fails, though this should be rare/impossible in most envs
		// Best effort for padding
		return r
	}
	return r
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
