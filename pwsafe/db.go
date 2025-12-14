// The database type for a Password Safe V3 database
// The db specification - https://github.com/pwsafe/pwsafe/blob/master/docs/formatV3.txt

package pwsafe

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	"github.com/pborman/uuid"
)

// V3 The type representing a password safe v3 database
type V3 struct {
	CBCIV         [16]byte //Random initial value for CBC
	EncryptionKey [32]byte
	Header        header
	HMAC          [32]byte //32bytes keyed-hash MAC with SHA-256 as the hash function.
	HMACKey       [32]byte
	Iter          uint32 //the number of iterations on the hash function to create the stretched key
	LastMod       time.Time
	LastSavePath  string
	Records       map[string]Record //the key is the record title
	Salt          [32]byte
	StretchedKey  [sha256.Size]byte
}

// NewV3 - create and initialize a new pwsafe.V3 db
func NewV3(name, password string) *V3 {
	var db V3
	db.Header = newHeader(name)
	db.Records = make(map[string]Record)

	// Set the password
	db.SetPassword(password)
	return &db
}

// DeleteRecord Removes a record from the db
func (db *V3) DeleteRecord(title string) {
	delete(db.Records, title)
	db.LastMod = time.Now()
}

// Equal compares the content of two V3 DBs except for LastSave fields and fields with transient or changing values.
func (db *V3) Equal(other *V3) (bool, error) {
	if matches, err := db.Header.Equal(other.Header); !matches || err != nil {
		return matches, err
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
	return db.Header.LastSave.Before(db.LastMod)
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

	if prs && record.CreateTime.IsZero() {
		record.CreateTime = oldRecord.CreateTime
	}

	if record.UUID == [16]byte{} {
		record.UUID = [16]byte(uuid.NewRandom().Array())
	}
	record.ModTime = now
	db.Records[record.Title] = record
	db.LastMod = now
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
