// The database type for a Password Safe V3 database
// The db specification - http://sourceforge.net/p/passwordsafe/code/HEAD/tree/trunk/pwsafe/pwsafe/docs/formatV3.txt

package pwsafe

import (
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"io"
	"sort"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"golang.org/x/crypto/twofish"
)

//Record The primary type for password DB entries
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

//V3 The type representing a password safe v3 database
type V3 struct {
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

//DB The interface representing the core functionality availble for any password database
type DB interface {
	//todo 	Encrypt(io.Writer) (int, err)
	Decrypt(io.Reader, string) (int, error)
	GetRecord(string) (Record, bool)
	Groups() []string
	List() []string
	ListByGroup(string) []string
}

// Using the db Salt and Iter along with the passwd calculate the stretch key
func (db *V3) calculateStretchKey(passwd string) {
	iterations := int(db.Iter)
	salted := append([]byte(passwd), db.Salt...)
	stretched := sha256.Sum256(salted)
	for i := 0; i < iterations; i++ {
		stretched = sha256.Sum256(stretched[:])
	}
	db.StretchedKey = stretched
}

//Decrypt Decrypts the data in the reader using the given password and populates the information into the db
func (db *V3) Decrypt(reader io.Reader, passwd string) (int, error) {
	var bytesRead int
	// The TAG is 4 ascii characters, should be "PWS3"
	tag := make([]byte, 4)
	_, err := reader.Read(tag)
	if err != nil || string(tag) != "PWS3" {
		return bytesRead, errors.New("File is not a valid Password Safe v3 file")
	}
	bytesRead += 4

	// Read the Salt
	salt := make([]byte, 32)
	readSize, err := reader.Read(salt)
	if err != nil || readSize != 32 {
		return bytesRead, errors.New("Error reading File, salt is invalid")
	}
	bytesRead += 32
	db.Salt = salt

	// Read iter
	iter := make([]byte, 4)
	readSize, err = reader.Read(iter)
	if err != nil || readSize != 4 {
		return bytesRead, errors.New("Error reading File, invalid iterations")
	}
	bytesRead += 4
	db.Iter = uint32(uint32(iter[0]) | uint32(iter[1])<<8 | uint32(iter[2])<<16 | uint32(iter[3])<<24)

	// Verify the password
	db.calculateStretchKey(passwd)
	readHash := make([]byte, sha256.Size)
	var keyHash [sha256.Size]byte
	readSize, err = reader.Read(readHash)
	copy(keyHash[:], readHash)
	if err != nil || readSize != sha256.Size || keyHash != sha256.Sum256(db.StretchedKey[:]) {
		return bytesRead, errors.New("Invalid Password")
	}

	//extract the encryption and hmac keys
	keyData := make([]byte, 64)
	readSize, err = reader.Read(keyData)
	if err != nil || readSize != 64 {
		return bytesRead, errors.New("Error reading encryption/HMAC keys")
	}
	bytesRead += 64
	db.extractKeys(keyData)

	cbciv := make([]byte, 16)
	readSize, err = reader.Read(cbciv)
	if err != nil || readSize != 16 {
		return bytesRead, errors.New("Error reading Initial CBC value")
	}
	bytesRead += 16
	db.CBCIV = cbciv

	// All following fields are encrypted with twofish in CBC mode until the EOF, find the EOF
	var encryptedDB []byte
	var encryptedSize int
	for {
		blockBytes := make([]byte, twofish.BlockSize)
		readSize, err = reader.Read(blockBytes)
		if err != nil || readSize != twofish.BlockSize {
			return bytesRead, errors.New("Error reading Encrypted Data, possibly no EOF found")
		}

		if string(blockBytes) == "PWS3-EOFPWS3-EOF" {
			bytesRead += twofish.BlockSize
			break
		} else {
			encryptedSize += readSize
			encryptedDB = append(encryptedDB, blockBytes...)
		}
	}
	bytesRead += encryptedSize

	if len(encryptedDB)%twofish.BlockSize != 0 {
		return bytesRead, errors.New("Error, encrypted data size is not a multiple of the block size")
	}

	// HMAC 32bytes keyed-hash MAC with SHA-256 as the hash function. Calculated over all unencryped db data
	hmac := make([]byte, 32)
	readSize, err = reader.Read(hmac)
	if err != nil || readSize != 32 {
		return bytesRead, errors.New("Error reading HMAC")
	}
	bytesRead += 32
	db.HMAC = hmac

	block, err := twofish.NewCipher(db.EncryptionKey)
	decrypter := cipher.NewCBCDecrypter(block, db.CBCIV)
	decryptedDB := make([]byte, encryptedSize) // The EOF and HMAC are after the encrypted section
	decrypter.CryptBlocks(decryptedDB, encryptedDB)

	//Parse the decrypted DB, first the header
	hdrSize, err := db.parseHeader(decryptedDB)
	if err != nil {
		return bytesRead, errors.New("Error parsing the unencrypted header - " + err.Error())
	}

	_, err = db.parseRecords(decryptedDB[hdrSize:])
	if err != nil {
		return bytesRead, errors.New("Error parsing the unencrypted records - " + err.Error())
	}

	return bytesRead, nil
}

// Pull EncryptionKey and HMAC key from the 64byte keyData
func (db *V3) extractKeys(keyData []byte) {
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

//GetRecord Returns a record from the db with the title matching the given String
func (db V3) GetRecord(title string) (Record, bool) {
	r, prs := db.Records[title]
	return r, prs
}

//Groups Returns an slice of strings which match all groups used by records in the DB
func (db V3) Groups() []string {
	groups := make([]string, 0, len(db.Records))
	groupSet := make(map[string]bool)
	for _, value := range db.Records {
		if _, prs := groupSet[value.Group]; !prs {
			groupSet[value.Group] = true
			groups = append(groups, value.Group)
		}
	}
	sort.Strings(groups)
	return groups
}

//List Returns the titles of all the records in the db.
func (db V3) List() []string {
	entries := make([]string, 0, len(db.Records))
	for key := range db.Records {
		entries = append(entries, key)
	}
	sort.Strings(entries)
	return entries
}

//ListByGroup Returns the list of record titles that have the given group.
func (db V3) ListByGroup(group string) []string {
	entries := make([]string, 0, len(db.Records))
	for key, value := range db.Records {
		if value.Group == group {
			entries = append(entries, key)
		}
	}
	sort.Strings(entries)
	return entries
}

// Parse the header of the decrypted DB returning the size of the Header and any error or nil
// beginning with the Version type field, and terminated by the 'END' type field. The version number
// and END fields are mandatory
func (db *V3) parseHeader(decryptedDB []byte) (int, error) {
	fieldStart := 0
	for {
		if fieldStart > len(decryptedDB) {
			return 0, errors.New("No END field found in DB header")
		}
		fieldLength := byteToInt(decryptedDB[fieldStart : fieldStart+4])
		btype := byteToInt(decryptedDB[fieldStart+4 : fieldStart+5])

		data := decryptedDB[fieldStart+5 : fieldStart+fieldLength+5]
		fieldStart += fieldLength + 5
		//The next field must start on a block boundary
		blockmod := fieldStart % twofish.BlockSize
		if blockmod != 0 {
			fieldStart += twofish.BlockSize - blockmod
		}
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
			db.Name = string(data)
		case 0x0a: //description
			db.Description = string(data)
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
			return 0, errors.New("Encountered unknown Header Field " + string(btype))
		}
	}
}

// Parse the records returning records length and error or nil
// The EOF string records end with is "PWS3-EOFPWS3-EOF"
func (db *V3) parseRecords(records []byte) (int, error) {
	recordStart := 0
	db.Records = make(map[string]Record)
	for recordStart < len(records) {
		recordLength, err := db.parseNextRecord(records[recordStart:])
		if err != nil {
			return recordStart, errors.New("Error parsing record - " + err.Error())
		}
		recordStart += recordLength
	}

	if recordStart > len(records) {
		return recordStart, errors.New("Encountered a record with invalid length")
	}
	return recordStart, nil
}

// Parse a single record from the given records []byte, return record size
// Individual records stop with an END filed and UUID, Title and Password fields are mandatory all others are optional
func (db *V3) parseNextRecord(records []byte) (int, error) {
	fieldStart := 0
	var record Record
	for {
		fieldLength := byteToInt(records[fieldStart : fieldStart+4])
		btype := byteToInt(records[fieldStart+4 : fieldStart+5])
		data := records[fieldStart+5 : fieldStart+fieldLength+5]
		fieldStart += fieldLength + 5
		//The next field must start on a block boundary
		blockmod := fieldStart % twofish.BlockSize
		if blockmod != 0 {
			fieldStart += twofish.BlockSize - blockmod
		}
		switch btype {
		case 0x01:
			record.UUID = data
		case 0x02:
			record.Group = string(data)
		case 0x03:
			record.Title = string(data)
		case 0x04:
			record.Username = string(data)
		case 0x05:
			record.Notes = string(data)
		case 0x06:
			record.Password = string(data)
		case 0x07:
			continue
		case 0x08:
			continue
		case 0x09:
			continue
		case 0x0a: // password expiry time
			continue
		case 0x0c:
			continue
		case 0x0d:
			record.URL = string(data)
		case 0x0e: //autotype
			continue
		case 0x0f: //password history
			continue
		case 0x10: //password policy
			continue
		case 0x11: //password expiry interval
			continue
		case 0x12: //run command
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
			db.Records[record.Title] = record
			return fieldStart + fieldLength, nil
		default:
			return fieldStart, errors.New("Encountered unknown Record Field type - " + string(btype))
		}
	}
}

func byteToInt(b []byte) int {
	bint := uint32(b[0])
	for i := 1; i < len(b); i++ {
		shift := uint(i) * 8
		bint = bint | uint32(b[i])<<shift
	}
	return int(bint)
}
