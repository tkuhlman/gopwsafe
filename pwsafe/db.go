// The database type for a Password Safe V3 database
// The db specification - http://sourceforge.net/p/passwordsafe/code/HEAD/tree/trunk/pwsafe/pwsafe/docs/formatV3.txt

package pwsafe

import (
	"crypto/sha256"
	"errors"
	"os"
	"time"
	//	"code.google.com/p/go.crypto/twofish"
	"code.google.com/p/go-uuid/uuid"
)

type Record struct {
	AccessTime      time.Time
	CreateTime      time.Time
	Group           string
	ModTime         time.Time
	Notes           string
	Password        string
	PasswordModTime string
	Title           string
	Username        string
	URL             string
	UUID            uuid.UUID
}

type PWSafeV3 struct {
	// Note not all of the Header information from the specification is implemented
	Name         string
	Description  string
	Iter         uint32 //the number of iterations on the hash function to create the stretched key
	LastSave     time.Time
	Records      map[string]Record //the key is the record title
	Salt         []byte            // should be 32 bytes
	UUID         uuid.UUID
	StretchedKey [sha256.Size]byte
	Version      string
}

type DB interface {
	List() []string
}

// Using the db Salt and Iter along with the passwd calculate the stretch key
func (db *PWSafeV3) CalculateStretchKey(passwd string) {
	iterations := int(db.Iter)
	salted := append([]byte(passwd), db.Salt...)
	stretched := sha256.Sum256(salted)
	for i := 0; i < iterations; i++ {
		stretched = sha256.Sum256(stretched[:])
	}
	db.StretchedKey = stretched
}

func (db PWSafeV3) List() []string {
	contents := make([]string, 0)
	return contents
}

func OpenPWSafe(dbPath string, passwd string) (DB, error) {
	db := PWSafeV3{}

	// Open the file
	f, err := os.Open(dbPath)
	if err != nil {
		return db, err
	}
	defer f.Close()

	// The TAG is 4 ascii characters, should be "PWS3"
	tag := make([]byte, 4)
	_, err = f.Read(tag)
	if err != nil || string(tag) != "PWS3" {
		return db, errors.New("File is not a valid Password Safe v3 file")
	}

	// Read the Salt
	salt := make([]byte, 32)
	saltSize, err := f.Read(salt)
	if err != nil || saltSize != 32 {
		return db, errors.New("Error reading File, salt is invalid")
	}
	db.Salt = salt

	// Read iter
	iter := make([]byte, 4)
	iterSize, err := f.Read(iter)
	if err != nil || iterSize != 4 {
		return db, errors.New("Error reading File, invalid iterations")
	}
	db.Iter = uint32(uint32(iter[0]) | uint32(iter[1])<<8 | uint32(iter[2])<<16 | uint32(iter[3])<<24)

	// Verify the password
	db.CalculateStretchKey(passwd)
	readHash := make([]byte, sha256.Size)
	var keyHash [sha256.Size]byte
	keySize, err := f.Read(readHash)
	copy(keyHash[:], readHash)
	if err != nil || keySize != sha256.Size || keyHash != sha256.Sum256(db.StretchedKey[:]) {
		return db, errors.New("Invalid Password")
	}

	return db, nil
}
