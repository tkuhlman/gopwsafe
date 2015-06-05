// The database type for a Password Safe V3 database
// The db specification - http://sourceforge.net/p/passwordsafe/code/HEAD/tree/trunk/pwsafe/pwsafe/docs/formatV3.txt

package pwsafe

import (
	"code.google.com/p/go-uuid/uuid"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"golang.org/x/crypto/twofish"
	"os"
	"time"
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
	Name          string
	CBCIV         []byte //16 bytes - Random initial value for CBC
	Description   string
	EncryptionKey []byte //32 bytes
	HMACKey       []byte //32 bytes
	Iter          uint32 //the number of iterations on the hash function to create the stretched key
	LastSave      time.Time
	Records       map[string]Record //the key is the record title
	Salt          []byte            // should be 32 bytes
	UUID          uuid.UUID
	StretchedKey  [sha256.Size]byte
	Version       string
}

type DB interface {
	List() []string
}

// Using the db Salt and Iter along with the passwd calculate the stretch key
func (db *PWSafeV3) calculateStretchKey(passwd string) {
	iterations := int(db.Iter)
	salted := append([]byte(passwd), db.Salt...)
	stretched := sha256.Sum256(salted)
	for i := 0; i < iterations; i++ {
		stretched = sha256.Sum256(stretched[:])
	}
	db.StretchedKey = stretched
}

// Pull EncryptionKey and HMAC key from the 64byte keyData
func (db *PWSafeV3) extractKeys(keyData []byte) {
	c, _ := twofish.NewCipher(db.StretchedKey[:])
	k1 := make([]byte, 16)
	c.Decrypt(k1, keyData[:16])
	k2 := make([]byte, 16)
	c.Decrypt(k2, keyData[16:32])
	db.EncryptionKey = append(k1, k2...)

	l1 := make([]byte, 16)
	c.Decrypt(l1, keyData[32:48])
	l2 := make([]byte, 16)
	c.Decrypt(l2, keyData[48:])
	db.HMACKey = append(l1, l2...)
}

func (db PWSafeV3) List() []string {
	entries := make([]string, len(db.Records))
	for key := range db.Records {
		entries = append(entries, key)
	}
	return entries
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
	readSize, err := f.Read(salt)
	if err != nil || readSize != 32 {
		return db, errors.New("Error reading File, salt is invalid")
	}
	db.Salt = salt

	// Read iter
	iter := make([]byte, 4)
	readSize, err = f.Read(iter)
	if err != nil || readSize != 4 {
		return db, errors.New("Error reading File, invalid iterations")
	}
	db.Iter = uint32(uint32(iter[0]) | uint32(iter[1])<<8 | uint32(iter[2])<<16 | uint32(iter[3])<<24)

	// Verify the password
	db.calculateStretchKey(passwd)
	readHash := make([]byte, sha256.Size)
	var keyHash [sha256.Size]byte
	readSize, err = f.Read(readHash)
	copy(keyHash[:], readHash)
	if err != nil || readSize != sha256.Size || keyHash != sha256.Sum256(db.StretchedKey[:]) {
		return db, errors.New("Invalid Password")
	}

	//extract the encryption and hmac keys
	keyData := make([]byte, 64)
	readSize, err = f.Read(keyData)
	if err != nil || readSize != 64 {
		return db, errors.New("Error reading encryption/HMAC keys")
	}
	db.extractKeys(keyData)

	cbciv := make([]byte, 16)
	readSize, err = f.Read(cbciv)
	if err != nil || readSize != 16 {
		return db, errors.New("Error reading Initial CBC value")
	}
	db.CBCIV = cbciv

	// All following fields are encrypted with twofish in CBC mode
	block, err := twofish.NewCipher(db.EncryptionKey)
	decrypter := cipher.NewCBCDecrypter(block, db.CBCIV)
	finfo, _ := f.Stat()
	remainingSize := int(finfo.Size() - 152)
	encryptedDB := make([]byte, remainingSize)
	readSize, err = f.Read(encryptedDB)
	if err != nil || readSize != remainingSize {
		return db, errors.New("Error reading Encrypted Data")
	}

	if len(encryptedDB)%twofish.BlockSize != 0 {
		return db, errors.New("Error, data size is not a multiple of the block size")
	}
	decryptedDB := make([]byte, remainingSize)

	decrypter.CryptBlocks(decryptedDB, encryptedDB)

	//Parse the decrypted DB

	return db, nil
}
