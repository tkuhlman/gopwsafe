package pwsafe

import (
	"crypto/sha256"
	"io"
)

//Encrypt Encrypt the data in the db building it up in memory then writing to the writer, returns bytesWritten, error
func (db *V3) Encrypt(writer io.Writer) (int, error) {
	var dbBytes []byte

	// Set unencrypted DB headers
	dbBytes = append(dbBytes, "PWS3"...)

	//db.stretchedKey is used to create a new encryptionKey

	//update the LastSave time in the DB

	//todo
	// generate and write salt

	//todo
	// Write iter

	// Add the stretchedKey
	stretchedSha := sha256.Sum256(db.stretchedKey[:])
	dbBytes = append(dbBytes, stretchedSha[:]...)

	//todo
	// calculate and write encryption key and hmac key

	//todo
	// calculate and write cbc initial value

	// encrypt the core of the db

	// todo
	// Calculate and write hmac

	// Write out the db
	// todo - skip the write until we have an actual valid db implemented
	return 0, nil
	//bytesWritten, err : writer.Write(dbBrytes)
	//return bytesWritten, err
}
