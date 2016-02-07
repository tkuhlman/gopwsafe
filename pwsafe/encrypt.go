package pwsafe

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math/rand"
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
	copy(db.Salt[:], generateRandomBytes(32))

	// Write iter
	db.Iter = 86000

	// Add the stretchedKey
	stretchedSha := sha256.Sum256(db.stretchedKey[:])
	dbBytes = append(dbBytes, stretchedSha[:]...)

	// re-calculate, encrypt and write encryption key and hmac key
	copy(db.encryptionKey[:], generateRandomBytes(32))
	copy(db.HMACKey[:], generateRandomBytes(32))
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

//generateRandomBytes Generates []byte of random data the specified length
// todo can this only operate in 8 bit increments? If so valid size%8 and fail appropriately.  I need some tests for it either way
func generateRandomBytes(size int) []byte {
	r := make([]byte, size)
	bytesRand := make([]byte, 8)
	for i := 0; i < size; i += 8 {
		binary.PutVarint(bytesRand, rand.Int63())
		for j, value := range bytesRand {
			r[i+j] = value
		}
	}
	return r
}
