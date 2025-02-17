// The database type for a Password Safe V3 database
// The db specification - https://github.com/pwsafe/pwsafe/blob/master/docs/formatV3.txt

package pwsafe

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/pborman/uuid"
)

// V3 The type representing a password safe v3 database
type V3 struct {
	CBCIV          [16]byte //Random initial value for CBC
	Description    string   `field:"0a"`
	EmptyGroups    []string `field:"11"`
	EncryptionKey  [32]byte
	Filters        string   `field:"0b"`
	HMAC           [32]byte //32bytes keyed-hash MAC with SHA-256 as the hash function.
	HMACKey        [32]byte
	Iter           uint32 //the number of iterations on the hash function to create the stretched key
	LastMod        time.Time
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
	StretchedKey   [sha256.Size]byte
	Tree           string   `field:"03"`
	UUID           [16]byte `field:"01"`
	Version        [2]byte  `field:"00"`
}

// NewV3 - create and initialize a new pwsafe.V3 db
func NewV3(name, password string) *V3 {
	var db V3
	db.Name = name
	// create the initial UUID
	db.UUID = [16]byte(uuid.NewRandom().Array())
	// Set the DB version
	db.Version = [2]byte{0x10, 0x03} // DB Format version 0x0310
	db.Records = make(map[string]Record, 0)

	// Set the password
	db.SetPassword(password)
	return &db
}

// DeleteRecord Removes a record from the db
func (db *V3) DeleteRecord(title string) {
	delete(db.Records, title)
	db.LastMod = time.Now()
}

// Equal returns true if the two dbs have the same data but not necessarily the same keys nor same LastSave time
func (db *V3) Equal(other *V3) (bool, error) {
	// todo should I compare version?
	skipHeaderFields := map[string]bool{"LastSave": true, "LastSaveBy": true, "UUID": true, "Version": true}
	// restrict comparison to fields with a field struct tag
	otherStruct := structs.New(other)
	for _, field := range mapByFieldTag(db) {
		if _, skip := skipHeaderFields[field.Name()]; skip {
			continue
		}
		if !reflect.DeepEqual(field.Value(), otherStruct.Field(field.Name()).Value()) {
			return false, fmt.Errorf("%v fields not equal, %v != %v", field.Name(), field.Value(), otherStruct.Field(field.Name()).Value())
		}
	}

	// compare records
	if len(db.List()) != len(other.List()) {
		return false, fmt.Errorf("record lengths don't match, %v != %v", len(db.List()), len(other.List()))
	}
	for _, title := range db.List() {
		equal, err := db.Records[title].Equal(other.Records[title], true)
		if !equal {
			return false, err
		}
	}
	return true, nil
}

// GetName returns the database name or if unset the filename
func (db *V3) GetName() string {
	if db.Name == "" {
		splits := strings.Split(db.LastSavePath, "/")
		return splits[len(splits)-1]
	}
	return db.Name
}

// Groups Returns an slice of strings which match all groups used by records in the DB
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

// Identical returns true if the two dbs have the same fields including the cryptographic keys
// note this doesn't check times and uuid's of the records
func (db *V3) Identical(other *V3) (bool, error) {
	equal, err := db.Equal(other)
	if !equal {
		return false, err
	}
	dbStruct := structs.New(*db)
	otherStruct := structs.New(other)
	// TODO add back in UUID, for some reason it is not being read correctly at times but the code needs lots of cleanup before it will be clear why
	skipHeaderFields := []string{"LastSaveBy", "Version"}
	encryptionFields := []string{"CBCIV", "EncryptionKey", "HMACKey", "Iter", "Salt", "StretchedKey"}
	checkFields := append(skipHeaderFields, encryptionFields...)
	for _, fieldName := range checkFields {
		if !reflect.DeepEqual(dbStruct.Field(fieldName).Value(), otherStruct.Field(fieldName).Value()) {
			return false, fmt.Errorf("%v fields not equal, %v != %v", fieldName, dbStruct.Field(fieldName).Value(), otherStruct.Field(fieldName).Value())
		}
	}

	return true, nil
}

// List Returns the titles of all the records in the db.
func (db V3) List() []string {
	entries := make([]string, 0, len(db.Records))
	for key := range db.Records {
		entries = append(entries, key)
	}
	sort.Strings(entries)
	return entries
}

// ListByGroup Returns the list of record titles that have the given group.
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

// NeedsSave Returns true if the db has unsaved modifiations
func (db V3) NeedsSave() bool {
	return db.LastSave.Before(db.LastMod)
}

// SetPassword Sets the password that will be used to encrypt the file on next save
func (db *V3) SetPassword(pw string) error {
	// First recalculate the Salt and set iter
	db.Iter = 86000
	if _, err := rand.Read(db.Salt[:]); err != nil {
		return err
	}
	db.calculateStretchKey(pw)
	db.LastMod = time.Now()
	return nil
}

// SetRecord Adds or updates a record in the db
func (db *V3) SetRecord(record Record) {
	now := time.Now()
	//detect if there have been changes and only update if needed
	oldRecord, prs := db.Records[record.Title]
	if prs {
		equal, _ := oldRecord.Equal(record, false)
		if equal {
			return
		}
	} else {
		record.CreateTime = now
	}

	if record.UUID == [16]byte{} {
		record.UUID = [16]byte(uuid.NewRandom().Array())
	}
	record.ModTime = now
	db.Records[record.Title] = record
	db.LastMod = now
	// todo add checking of db and record times to the tests
}

// calculateHMAC calculate and set db.HMAC for the unencrypted data using HMACKey
func (db *V3) calculateHMAC(unencrypted []byte) {
	hmacHash := hmac.New(sha256.New, db.HMACKey[:])
	hmacHash.Write(unencrypted)
	copy(db.HMAC[:], hmacHash.Sum(nil))
}

// calculateStretchKey Using the db Salt and Iter along with the passwd calculate the stretch key
func (db *V3) calculateStretchKey(passwd string) {
	iterations := int(db.Iter)
	salted := append([]byte(passwd), db.Salt[:]...)
	stretched := sha256.Sum256(salted)
	for i := 0; i < iterations; i++ {
		stretched = sha256.Sum256(stretched[:])
	}
	db.StretchedKey = stretched
}

// TODO I may be able to replaces this with, binary.BigEndian.Uint32 or similar
func byteToInt(b []byte) int {
	bint := uint32(b[0])
	for i := 1; i < len(b); i++ {
		shift := uint(i) * 8
		bint = bint | uint32(b[i])<<shift
	}
	return int(bint)
}
