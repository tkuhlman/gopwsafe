package pwsafe

import (
	"crypto/rand"
	"crypto/sha256"
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

	// generate and write salt
	_, err := rand.Read(db.Salt[:])
	if err != nil {
		return 0, err
	}

	// Write iter
	db.Iter = 86000

	// Add the stretchedKey
	stretchedSha := sha256.Sum256(db.stretchedKey[:])
	dbBytes = append(dbBytes, stretchedSha[:]...)

	// re-calculate, encrypt and write encryption key and hmac key
	_, err = rand.Read(db.encryptionKey[:])
	if err != nil {
		return 0, err
	}
	_, err = rand.Read(db.HMACKey[:])
	if err != nil {
		return 0, err
	}
	cipherTwoFish, _ := twofish.NewCipher(db.stretchedKey[:])
	for _, block := range [][]byte{db.encryptionKey[:16], db.encryptionKey[16:], db.HMACKey[:16], db.HMACKey[16:]} {
		encrypted := make([]byte, 16)
		cipherTwoFish.Encrypt(encrypted, block)
		dbBytes = append(dbBytes, encrypted...)
	}

	//todo
	// calculate and write cbc initial value

	// todo
	// write then encrypt the core of the db
	// todo I should look into ways the byte mapping for types in the header as records can be expressed so I can reuse it for both encrypting and decrypting
	//  ideally I add a field to the struct then to the byte mapping and both encrypt/decrypt both support it.

	// todo
	// Calculate and write hmac

	// Write out the db
	// todo - skip the write until we have an actual valid db implemented
	return 0, nil
	//bytesWritten, err : writer.Write(dbBrytes)
	//return bytesWritten, err
}
