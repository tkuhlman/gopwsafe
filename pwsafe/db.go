// The database type for a Password Safe V3 database
// The db specification - http://sourceforge.net/p/passwordsafe/code/HEAD/tree/trunk/pwsafe/pwsafe/docs/formatV3.txt

package pwsafe

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
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
	HMAC          []byte //32 bytes
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

// Parse the header of the decrypted DB returning the size of the Header and any error or nil
// beginning with the Version type field, and terminated by the 'END' type field. The version number
// and END fields are mandatory
func (db PWSafeV3) ParseHeader(decryptedDB []byte) (int, error) {
	fieldStart := 0
	for {
		if fieldStart > len(decryptedDB) {
			return 0, errors.New("No END field found in DB header")
		}
		var fieldLength int
		buf := bytes.NewReader(decryptedDB[fieldStart : fieldStart+4])
		_ = binary.Read(buf, binary.LittleEndian, &fieldLength)

		// todo - review http://golang.org/pkg/bytes/
		// I need to use a buffer to read the btype field and then conditionally read the data as string, timestamp, etc as needed.
		btype := decryptedDB[fieldStart+4 : fieldStart+5]
		data := decryptedDB[fieldStart+5 : fieldStart+fieldLength]
		switch btype {
		case 0x00: //version
			db.Version = string(data)
		case 0x01: //uuuid
			db.UUID = data
		case 0x02: //preferences
			continue
		case 0x03: //tree
			continue
		case 0x04: //timestamp
			continue
		case 0x05: //who last save
			continue
		case 0x06: //last save timestamp
			continue
		case 0x07: //last save user
			continue
		case 0x08: //last save host
			continue
		case 0x09: //DB name
			db.Name = data
		case 0x0a: //description
			db.Description = data
		case 0x0b: //filters
			continue
		case 0x0f: //recently used
			continue
		case 0x10: //password policy
			continue
		case 0x11: //Empty Groups
			continue
		case 0xff: //end
			return fieldStart + fieldLength, nil
		default:
			return 0, errors.New("Encountered unknown Header Field")
		}
		fieldStart += fieldLength
	}
}

// Parse the records returning records length and error or nil
// The EOF string records end with is "PWS3-EOFPWS3-EOF"
func (db PWSafeV3) ParseRecords(records []byte) (int, error) {
	recordStart := 0
	for {
		if recordStart+twofish.BlockSize == len(records) {
			if string(records[recordStart:]) == "PWS3-EOFPWS3-EOF" {
				return recordStart, nil
			} else {
				return recordStart, errors.New("Invalid EOF")
			}
		}
		if recordStart > len(records) {
			return recordStart, errors.New("No EOF found in records")
		}
		recordLength, err := db.ParseNextRecord(records[recordStart:])
		if err != nil {
			return recordStart, errors.New("Error parsing record")
		}
		recordStart += recordLength
	}
}

// Parse a single record from the given records []byte, return record size
// Individual records stop with an END filed and UUID, Title and Password fields are mandatory all others are optional
func (db PWSafeV3) ParseNextRecord(records []byte) (int, error) {
	fieldStart := 0
	var record Record
	for {
		var fieldLength int
		buf := bytes.NewReader(records[fieldStart : fieldStart+4])
		_ = binary.Read(buf, binary.LittleEndian, &fieldLength)

		btype := records[fieldStart+4 : fieldStart+5]
		data := records[fieldStart+5 : fieldStart+fieldLength]
		switch btype {
		case 0x01:
			record.UUID = data
		case 0x02:
			record.Group = data
		case 0x03:
			record.Title = data
		case 0x04:
			record.Username = data
		case 0x05:
			record.Notes = data
		case 0x06:
			record.Password = data
		case 0x07:
			record.CreateTime = data
		case 0x08:
			record.PasswordModTime = data
		case 0x09:
			record.AccessTime = data
		case 0x0a: // password expiry time
			continue
		case 0x0c:
			record.ModTime = data
		case 0x0d:
			record.URL = data
		case 0x0e: //autotype
			continue
		case 0x0f: //password history
			continue
		case 0x10: //password policy
			continue
		case 0x11: //password expiry interval
			continue
		case 0x13: //double click action
			continue
		case 0x14: //email
			continue
		case 0x15: //protected entry
			continue
		case 0x16: //own symbol
			continue
		case 0x17: //shift double click action
			continue
		case 0x18: //password policy name
			continue
		case 0xff: //end
			return fieldStart + fieldLength, nil
		default:
			return fieldStart, errors.New("Encountered unknown Header Field")
		}
		fieldStart += fieldLength
	}
	db.Records[record.Title] = record
	return fieldStart, nil
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

	//Parse the decrypted DB, first the header
	hdrSize, err = db.ParseHeader(decryptedDB)
	if err != nil {
		return db, errors.New("Error parsing the unencrypted header")
	}

	recordSize, err = db.ParseRecords(decryptedDB[hdrSize:])
	if err != nil {
		return db, errors.New("Error parsing the unencrypted records")
	}

	// HMAC 32bytes keyed-hash MAC with SHA-256 as the hash function. Calculated over all db data until the EOF string
	if len(decryptedDB[hdrSize+recordSize:]) != 32 {
		return db, errors.New("Error reading HMAC value")
	}
	db.HMAC = decryptedDB[hdrSize+recordSize:]

	return db, nil
}
