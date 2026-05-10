// The database type for a Password Safe V3 database
// The db specification - https://github.com/pwsafe/pwsafe/blob/master/docs/formatV3.txt

package pwsafe

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
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

// uuidKey returns the map key string for a record UUID.
func uuidKey(uuid [16]byte) string {
	return fmt.Sprintf("%x", uuid)
}

// DeleteRecord removes a record from the db by UUID hex string.
func (db *V3) DeleteRecord(uuidHex string) {
	delete(db.Records, uuidHex)
	db.LastMod = time.Now()
}

// Equal compares the content of two V3 DBs except for LastSave fields and fields with transient or changing values.
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

// List Returns the UUID hex keys of all records in the db.
func (db V3) List() []string {
	entries := make([]string, 0, len(db.Records))
	for key := range db.Records {
		entries = append(entries, key)
	}
	return entries
}

// ListByGroup Returns the UUID hex keys of records that have the given group.
func (db V3) ListByGroup(group string) []string {
	entries := make([]string, 0, len(db.Records))
	for key, rec := range db.Records {
		if rec.Group == group {
			entries = append(entries, key)
		}
	}
	return entries
}

// Search returns titles of records matching all whitespace-separated terms in query.
// When namesOnly is true only title and group are searched; otherwise username,
// URL, and notes are included. Password is never searched.
func (db V3) Search(query string, namesOnly bool) []string {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return db.List()
	}
	var results []string
	for key, rec := range db.Records {
		var hay string
		if namesOnly {
			hay = strings.ToLower(rec.Title + "\n" + rec.Group)
		} else {
			hay = strings.ToLower(rec.Title + "\n" + rec.Group + "\n" + rec.Username + "\n" + rec.URL + "\n" + rec.Notes)
		}
		match := true
		for _, t := range terms {
			if !strings.Contains(hay, t) {
				match = false
				break
			}
		}
		if match {
			results = append(results, key)
		}
	}
	return results
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

// SetRecord Adds or updates a record in the db, keyed by UUID. Returns the UUID hex key.
func (db *V3) SetRecord(record Record) string {
	now := time.Now()
	if record.UUID == [16]byte{} {
		record.UUID = [16]byte(uuid.NewRandom().Array())
	}
	key := uuidKey(record.UUID)
	oldRecord, prs := db.Records[key]
	if !prs {
		record.CreateTime = now
	} else if record.CreateTime.IsZero() {
		record.CreateTime = oldRecord.CreateTime
	}

	record.ModTime = now
	db.Records[key] = record
	db.LastMod = now
	return key
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
