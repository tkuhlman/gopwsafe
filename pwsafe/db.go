// The database type for a Password Safe V3 database
// The db specification - https://github.com/pwsafe/pwsafe/blob/master/docs/formatV3.txt

package pwsafe

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"sort"
	"strings"
	"time"
)

//Record The primary type for password DB entries
type Record struct {
	AccessTime             time.Time `field:"09"`
	Autotype               string    `field:"0e"`
	CreateTime             time.Time `field:"07"`
	DoubleClickAction      [2]byte   `field:"13"`
	Email                  string    `field:"14"`
	Group                  string    `field:"02"`
	ModTime                time.Time `field:"0c"`
	Notes                  string    `field:"05"`
	Password               string    `field:"06"`
	PasswordExpiry         time.Time `field:"0a"`
	PasswordExpiryInterval [4]byte   `field:"11"`
	PasswordHistory        string    `field:"0f"`
	PasswordModTime        string    `field:"08"`
	PasswordPolicy         string    `field:"10"`
	PasswordPolicyName     string    `field:"18"`
	ProtectedEntry         byte      `field:"15"`
	RunCommand             string    `field:"12"`
	ShiftDoubleClickAction [2]byte   `field:"17"`
	Title                  string    `field:"03"`
	Username               string    `field:"04"`
	URL                    string    `field:"0d"`
	UUID                   [16]byte  `field:"01"`
}

//V3 The type representing a password safe v3 database
type V3 struct {
	CBCIV          [16]byte //Random initial value for CBC
	Description    string   `field:"0a"`
	EmptyGroups    []string `field:"11"`
	encryptionKey  [32]byte
	Filters        string   `field:"0b"`
	HMAC           [32]byte //32bytes keyed-hash MAC with SHA-256 as the hash function.
	HMACKey        [32]byte
	Iter           uint32    //the number of iterations on the hash function to create the stretched key
	LastSave       time.Time `field:"04"`
	LastSaveBy     []byte    `field:"06"`
	LastSaveHost   []byte    `field:"08"`
	LastSavePath   string
	LastSaveUser   []byte            `field:"07"`
	Name           string            `field:"09"`
	PasswordPolicy string            `field:"10"`
	Preferences    string            `field:"02"`
	Records        map[string]Record //the key is the record title
	RecentyUsed    string            `field:"0f"`
	Salt           [32]byte
	stretchedKey   [sha256.Size]byte
	Tree           string   `field:"03"`
	UUID           [16]byte `field:"01"`
	Version        [2]byte  `field:"00"`
}

//DB The interface representing the core functionality availble for any password database
type DB interface {
	Encrypt(io.Writer) (int, error)
	Equal(*DB) bool
	Decrypt(io.Reader, string) (int, error)
	GetName() string
	GetRecord(string) (Record, bool)
	Groups() []string
	Identical(*DB) bool
	List() []string
	ListByGroup(string) []string
	//todo - Make sure to calculate initial UUID -	NewDB(string) *DB
	SetPassword(string) error
	SetRecord(Record)
	DeleteRecord(string)
}

//calculateHMAC calculate and set db.HMAC for the unencrypted data using HMACKey
func (db *V3) calculateHMAC(unencrypted []byte) {
	hmacHash := hmac.New(sha256.New, db.HMACKey[:])
	hmacHash.Write(unencrypted)
	copy(db.HMAC[:], hmacHash.Sum(nil))
}

//calculateStretchKey Using the db Salt and Iter along with the passwd calculate the stretch key
func (db *V3) calculateStretchKey(passwd string) {
	iterations := int(db.Iter)
	salted := append([]byte(passwd), db.Salt[:]...)
	stretched := sha256.Sum256(salted)
	for i := 0; i < iterations; i++ {
		stretched = sha256.Sum256(stretched[:])
	}
	db.stretchedKey = stretched
}

//DeleteRecord Removes a record from the db
func (db V3) DeleteRecord(title string) {
	delete(db.Records, title)
}

// GetName returns the database name or if unset the filename
func (db *V3) GetName() string {
	if db.Name == "" {
		splits := strings.Split(db.LastSavePath, "/")
		return splits[len(splits)-1]
	}
	return db.Name
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

//SetPassword Sets the password that will be used to encrypt the file on next save
func (db *V3) SetPassword(pw string) error {
	// First recalculate the Salt and set iter
	db.Iter = 86000
	if _, err := rand.Read(db.Salt[:]); err != nil {
		return err
	}
	db.calculateStretchKey(pw)
	return nil
}

//SetRecord Adds or updates a record in the db
func (db V3) SetRecord(record Record) {
	db.Records[record.Title] = record
}

func byteToInt(b []byte) int {
	bint := uint32(b[0])
	for i := 1; i < len(b); i++ {
		shift := uint(i) * 8
		bint = bint | uint32(b[i])<<shift
	}
	return int(bint)
}
