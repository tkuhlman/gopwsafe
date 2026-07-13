package pwsafe

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/* The test databases simple.dat and three.dat were made using Loxodo (https://github.com/sommer/loxodo)
Some other test dbs can be found at https://github.com/ronys/pypwsafe/tree/master/test_safes
these all have the password 'bogus12345'
*/

func TestKeys(t *testing.T) {
	var db V3
	db.Iter = 2048
	db.Salt = [32]byte{224, 70, 145, 8, 59, 173, 47, 241, 203, 157, 83, 209, 22, 55, 151, 157, 96, 234, 194, 167, 175, 251, 199, 145, 7, 219, 203, 168, 6, 166, 238, 241}
	expectedKey := [32]byte{243, 201, 143, 194, 139, 58, 186, 186, 133, 14, 238, 200, 139, 153, 45, 247, 215, 251, 24, 49, 28, 170, 157, 181, 21, 174, 129, 231, 234, 62, 51, 203}

	// tests the stretchedKey
	db.calculateStretchKey("password")
	assert.Equal(t, db.StretchedKey, expectedKey)

	keyBuf := &bytes.Buffer{}
	assert.NoError(t, db.refreshEncryptedKeys(keyBuf))
	createdEncryptionKey := db.EncryptionKey
	createdHMACKey := db.HMACKey

	// extract the keys from the encrypted bytes and compare to the original
	db.extractKeys(keyBuf.Bytes())
	assert.Equal(t, createdEncryptionKey, db.EncryptionKey)
	assert.Equal(t, createdHMACKey, db.HMACKey)
}

func TestInvalidFile(t *testing.T) {
	_, err := OpenPWSafeFile("./db.go", "password")
	assert.Equal(t, err, errors.New("file is not a valid Password Safe v3 file"))
	_, err = OpenPWSafeFile("./notafile", "password")
	assert.NotNil(t, err)
}

func TestSetRecordTimes(t *testing.T) {
	db := NewV3("test", "password")
	record := Record{Title: "Test Record", Password: "password"}

	// Test new record
	db.SetRecord(record)
	savedRecord, ok := db.Records["Test Record"]
	assert.True(t, ok)
	assert.False(t, savedRecord.CreateTime.IsZero())
	assert.False(t, savedRecord.ModTime.IsZero())
	assert.False(t, db.LastMod.IsZero())

	// Capture times
	createTime := savedRecord.CreateTime
	modTime := savedRecord.ModTime
	dbLastMod := db.LastMod

	// Sleep to ensure time difference
	time.Sleep(1 * time.Second)

	// Test update record
	record.Password = "newpassword"
	db.SetRecord(record)
	updatedRecord, ok := db.Records["Test Record"]
	assert.True(t, ok)
	assert.Equal(t, createTime, updatedRecord.CreateTime)
	assert.True(t, updatedRecord.ModTime.After(modTime))
	assert.True(t, db.LastMod.After(dbLastMod))
}

// TestSetRecordKeyedByUUIDSameTitleDifferentUUID is a regression test ensuring that,
// with KeyByUUID enabled, SetRecord treats records with the same Title but no matching
// UUID as distinct entries, rather than overwriting one another (the default behavior
// exercised by TestSetRecordTimes above, which stays keyed by Title for compatibility).
func TestSetRecordKeyedByUUIDSameTitleDifferentUUID(t *testing.T) {
	db := NewV3("test", "password")
	db.KeyByUUID = true

	db.SetRecord(Record{Title: "Google", Username: "user1", Password: "pass1"})
	db.SetRecord(Record{Title: "Google", Username: "user2", Password: "pass2"})

	assert.Equal(t, 2, len(db.Records), "both same-titled records should be kept as distinct entries")

	usernames := make([]string, 0, 2)
	for _, record := range db.Records {
		usernames = append(usernames, record.Username)
	}
	assert.ElementsMatch(t, []string{"user1", "user2"}, usernames)
}

// TestDeleteRecordByUUID verifies precise, identity-based deletion when KeyByUUID is
// enabled, which matters when multiple records share a Title and DeleteRecord's
// title-based match would be ambiguous.
func TestDeleteRecordByUUID(t *testing.T) {
	db := NewV3("test", "password")
	db.KeyByUUID = true

	db.SetRecord(Record{Title: "Google", Username: "user1", Password: "pass1"})
	db.SetRecord(Record{Title: "Google", Username: "user2", Password: "pass2"})
	assert.Equal(t, 2, len(db.Records))

	var target Record
	for _, record := range db.Records {
		if record.Username == "user2" {
			target = record
		}
	}

	db.DeleteRecordByUUID(target.UUID)
	assert.Equal(t, 1, len(db.Records))

	remaining, ok := db.RecordByTitle("Google")
	assert.True(t, ok)
	assert.NotEqual(t, target.UUID, remaining.UUID)
}
