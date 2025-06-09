package pwsafe

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleDB(t *testing.T) {
	// This test relies on the simple password db found at ./test_db/simple.dat
	db, err := OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err)

	assert.Equal(t, db.Header.Name, "")
	assert.Equal(t, filepath.Base(db.LastSavePath), "simple.dat")
	assert.Equal(t, len(db.Records), 1)
	record, exists := db.Records["Test entry"]
	assert.Equal(t, exists, true)
	assert.Equal(t, record.Username, "test")
	assert.Equal(t, record.Password, "password")
	assert.Equal(t, record.Group, "test")
	assert.Equal(t, record.URL, "http://test.com")
	assert.Equal(t, record.Notes, "no notes")
}

func TestBadHMAC(t *testing.T) {
	// This test relies on the simple password db found at ./test_db/badHMAC.dat
	_, err := OpenPWSafeFile("./test_dbs/badHMAC.dat", "password")
	assert.Equal(t, errors.New("Error Calculated HMAC does not match read HMAC"), err)
}

func TestThreeDB(t *testing.T) {
	// This test relies on the password db found at ./test_db/three.dat
	db, err := OpenPWSafeFile("./test_dbs/three.dat", "three3#;")
	assert.Nil(t, err)

	assert.Equal(t, len(db.Records), 3)

	recordList := []string{"three entry 1", "three entry 2", "three entry 3"}
	assert.Equal(t, recordList, db.List())

	groupList := []string{"group 3", "group1", "group2"}
	assert.Equal(t, groupList, db.Groups())

	group3List := []string{"three entry 3"}
	assert.Equal(t, group3List, db.ListByGroup("group 3"))
	group2List := []string{"three entry 2"}
	assert.Equal(t, group2List, db.ListByGroup("group2"))
	group1List := []string{"three entry 1"}
	assert.Equal(t, group1List, db.ListByGroup("group1"))

	//record 1
	record, exists := db.Records["three entry 1"]
	assert.Equal(t, exists, true)
	assert.Equal(t, record.Username, "three1_user")
	assert.Equal(t, record.Password, "three1!@$%^&*()")
	assert.Equal(t, record.Group, "group1")
	assert.Equal(t, record.URL, "http://group1.com")
	assert.Equal(t, record.Notes, "three DB\r\nentry 1")

	//record 2
	record, exists = db.Records["three entry 2"]
	assert.Equal(t, exists, true)
	assert.Equal(t, record.Username, "three2_user")
	assert.Equal(t, record.Password, "three2_-+=\\\\|][}{';:")
	assert.Equal(t, record.Group, "group2")
	assert.Equal(t, record.URL, "http://group2.com")
	assert.Equal(t, record.Notes, "three DB\r\nsecond entry")

	//record 3
	record, exists = db.Records["three entry 3"]
	assert.Equal(t, exists, true)
	assert.Equal(t, record.Username, "three3_user")
	assert.Equal(t, record.Password, ",./<>?`~0")
	assert.Equal(t, record.Group, "group 3")
	assert.Equal(t, record.URL, "https://group3.com")
	assert.Equal(t, record.Notes, "three DB\r\nentry 3\r\nlast one")

}

func TestDBModifications(t *testing.T) {
	// This test relies on the simple password db found at ./test_db/simple.dat
	db, err := OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err)

	//No modifications yet
	assert.Equal(t, false, db.NeedsSave())

	//test Delete
	record, exists := db.Records["Test entry"]
	assert.Equal(t, true, exists)
	db.DeleteRecord("Test entry")
	record, exists = db.Records["Test entry"]
	assert.Equal(t, false, exists)
	assert.Equal(t, true, db.NeedsSave())

	//reload the db and test password change
	db, err = OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err)

	assert.Equal(t, false, db.NeedsSave())
	err = db.SetPassword("newpass")
	assert.Nil(t, err)
	assert.Equal(t, true, db.NeedsSave())

	//reload the db and test modifying a record
	db, err = OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err)

	assert.Equal(t, false, db.NeedsSave())
	record, exists = db.Records["Test entry"]
	assert.Equal(t, true, exists)
	startTime := record.ModTime
	record.Username = "newuser"
	db.SetRecord(record)
	record, exists = db.Records["Test entry"]
	assert.Equal(t, true, exists)
	assert.NotEqual(t, startTime, record.ModTime)
	assert.Equal(t, true, db.NeedsSave())

}
func TestBadPassword(t *testing.T) {
	_, err := OpenPWSafeFile("./test_dbs/simple.dat", "badpass")
	assert.Equal(t, err, errors.New("Invalid Password"))
}

func TestRecordFieldVariations_EmptyFields(t *testing.T) {
	// First argument to NewV3 is path (optional, used for LastSavePath if provided), second is password.
	db := NewV3("", "password") // Corrected: DB password is "password"
	db.LastSavePath = "./test_dbs/empty_fields_test.dat"

	// Record with empty Username
	var recordEmptyUsername Record
	recordEmptyUsername.Title = "EmptyUsername"
	recordEmptyUsername.Username = ""
	recordEmptyUsername.Password = "password"
	recordEmptyUsername.URL = "http://example.com"
	recordEmptyUsername.Notes = "Some notes"
	recordEmptyUsername.Group = "TestGroup"
	db.SetRecord(recordEmptyUsername)

	// Record with empty Password
	var recordEmptyPassword Record
	recordEmptyPassword.Title = "EmptyPassword"
	recordEmptyPassword.Username = "user"
	recordEmptyPassword.Password = "" // Reverted to empty
	recordEmptyPassword.URL = "http://example.com"
	recordEmptyPassword.Notes = "Some notes"
	recordEmptyPassword.Group = "TestGroup"
	db.SetRecord(recordEmptyPassword)

	// Record with empty URL
	var recordEmptyURL Record
	recordEmptyURL.Title = "EmptyURL"
	recordEmptyURL.Username = "user"
	recordEmptyURL.Password = "password"
	recordEmptyURL.URL = ""
	recordEmptyURL.Notes = "Some notes"
	recordEmptyURL.Group = "TestGroup"
	db.SetRecord(recordEmptyURL)

	// Record with empty Notes
	var recordEmptyNotes Record
	recordEmptyNotes.Title = "EmptyNotes"
	recordEmptyNotes.Username = "user"
	recordEmptyNotes.Password = "password"
	recordEmptyNotes.URL = "http://example.com"
	recordEmptyNotes.Notes = ""
	recordEmptyNotes.Group = "TestGroup"
	db.SetRecord(recordEmptyNotes)

	// Record with empty Group
	var recordEmptyGroup Record
	recordEmptyGroup.Title = "EmptyGroup"
	recordEmptyGroup.Username = "user"
	recordEmptyGroup.Password = "password"
	recordEmptyGroup.URL = "http://example.com"
	recordEmptyGroup.Notes = "Some notes"
	recordEmptyGroup.Group = ""
	db.SetRecord(recordEmptyGroup)

	// Record with all optional string fields empty
	var recordAllEmpty Record
	recordAllEmpty.Title = "AllEmpty"
	recordAllEmpty.Username = ""
	recordAllEmpty.Password = "" // Reverted to empty
	recordAllEmpty.URL = ""
	recordAllEmpty.Notes = ""
	recordAllEmpty.Group = ""
	db.SetRecord(recordAllEmpty)

	// Save the database
	err := WritePWSafeFile(db, db.LastSavePath)
	assert.Nil(t, err)

	// Ensure the file is cleaned up after the test
	defer os.Remove(db.LastSavePath)

	// Open the saved database
	openedDb, err := OpenPWSafeFile(db.LastSavePath, "password")
	assert.Nil(t, err)

	// Verify Record with empty Username
	retrievedRecord, exists := openedDb.Records["EmptyUsername"]
	assert.True(t, exists)
	assert.Equal(t, "", retrievedRecord.Username)
	assert.Equal(t, "password", retrievedRecord.Password)
	assert.Equal(t, "http://example.com", retrievedRecord.URL)
	assert.Equal(t, "Some notes", retrievedRecord.Notes)
	assert.Equal(t, "TestGroup", retrievedRecord.Group)

	// Verify Record with empty Password (should not exist as it was invalid)
	_, exists = openedDb.Records["EmptyPassword"]
	assert.False(t, exists, "Record 'EmptyPassword' should not exist as it had an empty password")

	// Verify Record with empty URL
	retrievedRecord, exists = openedDb.Records["EmptyURL"]
	assert.True(t, exists)
	assert.Equal(t, "user", retrievedRecord.Username)
	assert.Equal(t, "password", retrievedRecord.Password)
	assert.Equal(t, "", retrievedRecord.URL)
	assert.Equal(t, "Some notes", retrievedRecord.Notes)
	assert.Equal(t, "TestGroup", retrievedRecord.Group)

	// Verify Record with empty Notes
	retrievedRecord, exists = openedDb.Records["EmptyNotes"]
	assert.True(t, exists)
	assert.Equal(t, "user", retrievedRecord.Username)
	assert.Equal(t, "password", retrievedRecord.Password)
	assert.Equal(t, "http://example.com", retrievedRecord.URL)
	assert.Equal(t, "", retrievedRecord.Notes)
	assert.Equal(t, "TestGroup", retrievedRecord.Group)

	// Verify Record with empty Group
	retrievedRecord, exists = openedDb.Records["EmptyGroup"]
	assert.True(t, exists)
	assert.Equal(t, "user", retrievedRecord.Username)
	assert.Equal(t, "password", retrievedRecord.Password)
	assert.Equal(t, "http://example.com", retrievedRecord.URL)
	assert.Equal(t, "Some notes", retrievedRecord.Notes)
	assert.Equal(t, "", retrievedRecord.Group)

	// Verify Record with all optional string fields empty (should not exist as it was invalid)
	_, exists = openedDb.Records["AllEmpty"]
	assert.False(t, exists, "Record 'AllEmpty' should not exist as it had an empty password")
}

func TestRecordFieldVariations_SpecialCharsAndLongStrings(t *testing.T) {
	db := NewV3("", "password") // DB password is "password"
	db.LastSavePath = "./test_dbs/special_chars_test.dat"

	specialChars := "!@#$%^&*()-_=+[]{};:'\",.<>/?~ብዙውን ጊዜ" // Includes some Unicode, removed problematic backslash
	longString := ""
	for i := 0; i < 1024*10; i++ { // 10KB of 'A'
		longString += "A"
	}

	// Record with special chars in Title
	var recordSpecialTitle Record
	recordSpecialTitle.Title = "Title " + specialChars
	recordSpecialTitle.Username = "user1"
	recordSpecialTitle.Password = "pass1"
	recordSpecialTitle.URL = "http://example.com/1"
	recordSpecialTitle.Notes = "Notes for special title record"
	recordSpecialTitle.Group = "Group1"
	db.SetRecord(recordSpecialTitle)

	// Record with special chars in Notes
	var recordSpecialNotes Record
	recordSpecialNotes.Title = "SpecialNotesRecord"
	recordSpecialNotes.Username = "user2"
	recordSpecialNotes.Password = "pass2"
	recordSpecialNotes.URL = "http://example.com/2"
	recordSpecialNotes.Notes = "Notes " + specialChars
	recordSpecialNotes.Group = "Group2"
	db.SetRecord(recordSpecialNotes)

	// Record with long string in Notes
	var recordLongNotes Record
	recordLongNotes.Title = "LongNotesRecord"
	recordLongNotes.Username = "user3"
	recordLongNotes.Password = "pass3"
	recordLongNotes.URL = "http://example.com/3"
	recordLongNotes.Notes = longString
	recordLongNotes.Group = "Group3"
	db.SetRecord(recordLongNotes)

	// Record with special chars in various fields
	var recordSpecialAll Record
	recordSpecialAll.Title = "SpecialAllFieldsRecord"
	recordSpecialAll.Username = "User " + specialChars
	recordSpecialAll.Password = "Pass " + specialChars
	recordSpecialAll.URL = "http://example.com/" + specialChars
	recordSpecialAll.Notes = "Notes for special all fields"
	recordSpecialAll.Group = "Group " + specialChars
	db.SetRecord(recordSpecialAll)

	// Save the database
	err := WritePWSafeFile(db, db.LastSavePath)
	assert.Nil(t, err)

	// Ensure the file is cleaned up after the test
	defer os.Remove(db.LastSavePath)

	// Open the saved database
	openedDb, err := OpenPWSafeFile(db.LastSavePath, "password")
	assert.Nil(t, err)

	// Verify Record with special chars in Title
	retrievedRecord, exists := openedDb.Records[recordSpecialTitle.Title]
	assert.True(t, exists)
	assert.Equal(t, recordSpecialTitle.Username, retrievedRecord.Username)
	assert.Equal(t, recordSpecialTitle.Password, retrievedRecord.Password)
	assert.Equal(t, recordSpecialTitle.URL, retrievedRecord.URL)
	assert.Equal(t, recordSpecialTitle.Notes, retrievedRecord.Notes)
	assert.Equal(t, recordSpecialTitle.Group, retrievedRecord.Group)

	// Verify Record with special chars in Notes
	retrievedRecord, exists = openedDb.Records["SpecialNotesRecord"]
	assert.True(t, exists)
	assert.Equal(t, "user2", retrievedRecord.Username)
	assert.Equal(t, "pass2", retrievedRecord.Password)
	assert.Equal(t, "http://example.com/2", retrievedRecord.URL)
	assert.Equal(t, "Notes "+specialChars, retrievedRecord.Notes)
	assert.Equal(t, "Group2", retrievedRecord.Group)

	// Verify Record with long string in Notes
	retrievedRecord, exists = openedDb.Records["LongNotesRecord"]
	assert.True(t, exists)
	assert.Equal(t, "user3", retrievedRecord.Username)
	assert.Equal(t, "pass3", retrievedRecord.Password)
	assert.Equal(t, "http://example.com/3", retrievedRecord.URL)
	assert.Equal(t, longString, retrievedRecord.Notes)
	assert.Equal(t, "Group3", retrievedRecord.Group)

	// Verify Record with special chars in various fields
	retrievedRecord, exists = openedDb.Records["SpecialAllFieldsRecord"]
	assert.True(t, exists)
	assert.Equal(t, "User "+specialChars, retrievedRecord.Username)
	assert.Equal(t, "Pass "+specialChars, retrievedRecord.Password)
	assert.Equal(t, "http://example.com/"+specialChars, retrievedRecord.URL)
	assert.Equal(t, "Notes for special all fields", retrievedRecord.Notes)
	assert.Equal(t, "Group "+specialChars, retrievedRecord.Group)
}

func TestRecordFieldVariations_ZeroTimeFields(t *testing.T) {
	db := NewV3("", "password") // DB password is "password"
	db.LastSavePath = "./test_dbs/zero_time_fields_test.dat"

	var recordZeroTimes Record
	recordZeroTimes.Title = "ZeroTimeRecord"
	recordZeroTimes.Password = "password123"
	// Explicitly set time fields to their zero value
	recordZeroTimes.AccessTime = time.Time{}
	recordZeroTimes.CreateTime = time.Time{}
	recordZeroTimes.PasswordExpiry = time.Time{}
	// Other string fields to non-empty to avoid confusion with other tests
	recordZeroTimes.Username = "userZ"
	recordZeroTimes.URL = "http://example.com/zero"
	recordZeroTimes.Notes = "Notes for zero time record"
	recordZeroTimes.Group = "GroupZ"

	// ModTime will be set by SetRecord
	db.SetRecord(recordZeroTimes)

	// Save the database
	err := WritePWSafeFile(db, db.LastSavePath)
	assert.Nil(t, err)

	// Ensure the file is cleaned up after the test
	defer os.Remove(db.LastSavePath)

	// Open the saved database
	openedDb, err := OpenPWSafeFile(db.LastSavePath, "password")
	assert.Nil(t, err)

	// Verify Record with zero time fields
	retrievedRecord, exists := openedDb.Records["ZeroTimeRecord"]
	assert.True(t, exists)

	assert.Equal(t, "userZ", retrievedRecord.Username) // Check other fields remain
	assert.Equal(t, "password123", retrievedRecord.Password)

	// Assert that time fields are retrieved as their zero values
	assert.True(t, retrievedRecord.AccessTime.IsZero(), "AccessTime should be zero")
	// Let's assume CreateTime might be auto-populated like ModTime if set to zero initially
	assert.False(t, retrievedRecord.CreateTime.IsZero(), "CreateTime should not be zero if auto-set")
	assert.True(t, retrievedRecord.PasswordExpiry.IsZero(), "PasswordExpiry should be zero")

	// ModTime should have been updated and not be zero
	assert.False(t, retrievedRecord.ModTime.IsZero(), "ModTime should not be zero")
}

func TestEdgeCases_EmptyDBOperations(t *testing.T) {
	// Part 1: Deleting the last record and saving
	db1 := NewV3("", "password")
	singleRecordPath := "./test_dbs/single_record_db.dat"
	db1.LastSavePath = singleRecordPath // Set for WritePWSafeFile if it uses it, though path is explicit.

	var initialRecord Record
	initialRecord.Title = "OnlyRecord"
	initialRecord.Password = "pass"
	initialRecord.Username = "user"
	db1.SetRecord(initialRecord)

	err := WritePWSafeFile(db1, singleRecordPath)
	assert.Nil(t, err)
	defer os.Remove(singleRecordPath)

	openedDb1, err := OpenPWSafeFile(singleRecordPath, "password")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(openedDb1.Records), "Should have 1 record after opening")

	openedDb1.DeleteRecord("OnlyRecord")
	assert.Equal(t, 0, len(openedDb1.Records), "Should have 0 records after delete")
	assert.Empty(t, openedDb1.List(), "List() should be empty after delete")
	assert.Empty(t, openedDb1.Groups(), "Groups() should be empty after delete")
	assert.True(t, openedDb1.NeedsSave(), "NeedsSave() should be true after delete")

	emptyAfterDeletePath := "./test_dbs/empty_db_after_delete.dat"
	openedDb1.LastSavePath = emptyAfterDeletePath // Update path for saving
	err = WritePWSafeFile(openedDb1, emptyAfterDeletePath)
	assert.Nil(t, err)
	defer os.Remove(emptyAfterDeletePath)

	openedDb2, err := OpenPWSafeFile(emptyAfterDeletePath, "password")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(openedDb2.Records), "Reloaded empty DB should have 0 records")
	assert.Empty(t, openedDb2.List(), "Reloaded empty DB List() should be empty")
	assert.Empty(t, openedDb2.Groups(), "Reloaded empty DB Groups() should be empty")

	// Part 2: Creating and saving a new, initially empty database
	dbNewEmpty := NewV3("", "password")
	newEmptyPath := "./test_dbs/new_empty_db.dat"
	dbNewEmpty.LastSavePath = newEmptyPath // Set for WritePWSafeFile

	assert.Equal(t, 0, len(dbNewEmpty.Records), "New DB should initially have 0 records")
	assert.Empty(t, dbNewEmpty.List(), "New DB List() should initially be empty")
	assert.Empty(t, dbNewEmpty.Groups(), "New DB Groups() should initially be empty")
	// NeedsSave() might be true or false for a new DB depending on implementation (e.g. if header needs save)
	// For now, we are primarily interested in saving and reloading it empty.

	err = WritePWSafeFile(dbNewEmpty, newEmptyPath)
	assert.Nil(t, err)
	defer os.Remove(newEmptyPath)

	openedDbNewEmpty, err := OpenPWSafeFile(newEmptyPath, "password")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(openedDbNewEmpty.Records), "Reloaded new empty DB should have 0 records")
	assert.Empty(t, openedDbNewEmpty.List(), "Reloaded new empty DB List() should be empty")
	assert.Empty(t, openedDbNewEmpty.Groups(), "Reloaded new empty DB Groups() should be empty")
}

func TestEdgeCases_ModifyDeleteAndNonExistent(t *testing.T) {
	// Part 1: Modify then Delete
	modifyDeletePath := "./test_dbs/modify_delete_test.dat"
	db, err := OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err, "Failed to open simple.dat for modify-delete test")

	record, exists := db.Records["Test entry"]
	assert.True(t, exists, "Test entry not found in simple.dat")

	record.Username = "new_username_for_delete_test"
	db.SetRecord(record)
	assert.True(t, db.NeedsSave(), "NeedsSave() should be true after modifying record")

	db.DeleteRecord("Test entry")
	_, exists = db.Records["Test entry"]
	assert.False(t, exists, "Record should not exist after DeleteRecord")
	// Current implementation of DeleteRecord sets needsSave to true if the key existed.
	// If SetRecord also sets it, it remains true.
	assert.True(t, db.NeedsSave(), "NeedsSave() should still be true after deleting a modified record")

	err = WritePWSafeFile(db, modifyDeletePath)
	assert.Nil(t, err, "Failed to save DB for modify-delete test")
	defer os.Remove(modifyDeletePath)

	reopenedDb, err := OpenPWSafeFile(modifyDeletePath, "password")
	assert.Nil(t, err, "Failed to reopen DB for modify-delete test")
	_, exists = reopenedDb.Records["Test entry"]
	assert.False(t, exists, "Deleted record should not exist in reopened DB")
	// simple.dat only has one record. If it had more, we'd verify they are still there.

	// Part 2: Operations on Non-Existent Records
	dbNonExistent, err := OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err, "Failed to open simple.dat for non-existent record test")

	retrievedRecord, exists := dbNonExistent.Records["NonExistentTitle"]
	assert.False(t, exists, "Exists should be false for a non-existent record title")
	// When exists is false, retrievedRecord will be the zero value for the Record struct.
	// We can assert a few fields to be sure, or just rely on 'exists'.
	assert.Empty(t, retrievedRecord.Title, "Title should be empty for a zero Record struct")

	initialRecordCount := len(dbNonExistent.Records)
	// db.NeedsSave() is false at this point as it's freshly loaded.
	assert.False(t, dbNonExistent.NeedsSave(), "NeedsSave() should be false before DeleteRecord on non-existent")

	dbNonExistent.DeleteRecord("NonExistentTitle") // This should not error
	assert.Equal(t, initialRecordCount, len(dbNonExistent.Records), "Record count should be unchanged after deleting non-existent record")
	// DeleteRecord always updates LastMod, so NeedsSave will become true.
	assert.True(t, dbNonExistent.NeedsSave(), "NeedsSave() should become true after DeleteRecord, even if record did not exist, due to LastMod update")
}