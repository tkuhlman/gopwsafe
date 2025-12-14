package pwsafe

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSaveSimpleDB - save simple DB, reopen and verify contents match the original but keys don't
func TestSaveSimpleDB(t *testing.T) {
	// This test relies on the simple password db found at ./test_db/simple.dat
	source, err := OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err)

	//Set a new password, save a copy, open it and compare with the source
	source.SetPassword("passwordcopy")
	copyPath := "./test_dbs/simple-copy.dat"
	err = WritePWSafeFile(source, copyPath)
	defer os.Remove(copyPath)
	assert.Nil(t, err)
	dest, err := OpenPWSafeFile("./test_dbs/simple-copy.dat", "passwordcopy")
	assert.Nil(t, err)

	equal, err := source.Equal(dest)
	assert.NoError(t, err)
	assert.True(t, equal)

	// Reopen the original and verify keys have changed but content is the same
	orig, err := OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err)

	// On write dest gets version set but orig doesn't have it so just set to the same here
	orig.Header.Version = dest.Header.Version
	orig.Header.UUID = dest.Header.UUID
	// I expect the stretchedkey, salt, encryption key, hmac key and CBCIV to have changed
	// iter changes also but won't necessarily always.
	equal, err = orig.Equal(dest)
	assert.Nil(t, err)
	assert.Equal(t, true, equal)
}

func TestSaveEmptyDB(t *testing.T) {
	newDB := NewV3("", "password") // Path is empty, password is "password"
	newDB.Header.Name = "TestEmptyDB"
	emptySavePath := "./test_dbs/empty_new_save_test.dat"
	newDB.LastSavePath = emptySavePath // Though WritePWSafeFile takes path explicitly

	err := WritePWSafeFile(newDB, emptySavePath)
	assert.Nil(t, err)
	defer os.Remove(emptySavePath)

	loadedDB, err := OpenPWSafeFile(emptySavePath, "password")
	assert.Nil(t, err)

	assert.Equal(t, 0, len(loadedDB.Records), "Opened empty DB should have 0 records")
	assert.Equal(t, "TestEmptyDB", loadedDB.Header.Name, "Header.Name mismatch")

	// For Equal check, align fields modified by WritePWSafeFile
	// WritePWSafeFile updates Header.Version and Header.LastSave
	// It does not change UUID of an existing DB struct, and for a new DB, NewV3 sets it.
	newDB.Header.Version = loadedDB.Header.Version
	newDB.Header.LastSave = loadedDB.Header.LastSave
	// UUID should be the same as newDB's UUID was set by NewV3 and WritePWSafeFile doesn't change it.

	equal, err := newDB.Equal(loadedDB)
	assert.Nil(t, err)
	assert.True(t, equal, "Original empty DB and loaded empty DB should be equal after aligning version and lastSave")
}

func TestConsecutiveSaves(t *testing.T) {
	originalPath := "./test_dbs/simple.dat"
	save1Path := "./test_dbs/consecutive_save_1.dat"
	save2Path := "./test_dbs/consecutive_save_2.dat"

	// 1. Load an existing test database
	origDb, err := OpenPWSafeFile(originalPath, "password")
	assert.Nil(t, err, "Failed to open original simple.dat")

	// 2. Save it immediately to a new temporary file
	// WritePWSafeFile will update origDb.Header.LastSave and origDb.Header.Version in-memory
	// So, to compare later, we might need a deep copy or reload origDb if we want to compare to pristine state.
	// For now, let's use the potentially modified origDb for the first comparison.
	err = WritePWSafeFile(origDb, save1Path)
	assert.Nil(t, err)
	defer os.Remove(save1Path)

	// 3. Open this newly saved file
	loadedDb1, err := OpenPWSafeFile(save1Path, "password")
	assert.Nil(t, err)

	// 4. Compare the just-opened DB with the original simple.dat
	//    origDb's Header.Version and Header.LastSave were updated by WritePWSafeFile.
	//    loadedDb1 has these new values. So they should be equal.
	//    UUID should not have changed.
	equal, err := origDb.Equal(loadedDb1)
	assert.Nil(t, err)
	assert.True(t, equal, "origDb (after save) and loadedDb1 should be equal")

	// 5. Make a change to loadedDb1
	var newRecord Record
	newRecord.Title = "NewConsecutiveRecord"
	newRecord.Username = "consecutive_user"
	newRecord.Password = "consecutive_pass"
	newRecord.Group = "ConsecutiveGroup"
	loadedDb1.SetRecord(newRecord) // This updates loadedDb1.LastMod

	// Keep a snapshot of loadedDb1's relevant state before the next save for later comparison
	// For Equal(), we need to align Version and LastSave after the next save.
	// The records map and other header fields in loadedDb1 are what we want to compare.
	// Let's create a temporary V3 struct or simply rely on direct field checks for newRecord.

	// 6. Save it again to a *different* temporary file
	// This will update loadedDb1.Header.Version and loadedDb1.Header.LastSave in memory
	err = WritePWSafeFile(loadedDb1, save2Path)
	assert.Nil(t, err)
	defer os.Remove(save2Path)

	// 7. Open consecutive_save_2.dat
	loadedDb2, err := OpenPWSafeFile(save2Path, "password")
	assert.Nil(t, err)

	// 8. Verify that the changes made in step 2e are present
	retrievedNewRecord, exists := loadedDb2.Records["NewConsecutiveRecord"]
	assert.True(t, exists, "NewConsecutiveRecord not found in loadedDb2")
	assert.Equal(t, "consecutive_user", retrievedNewRecord.Username)
	assert.Equal(t, "consecutive_pass", retrievedNewRecord.Password)
	assert.Equal(t, "ConsecutiveGroup", retrievedNewRecord.Group)

	// 9. Compare loadedDb2 with the state of loadedDb1 (before it was saved to save2Path)
	//    loadedDb1's Header.Version and Header.LastSave were updated by the WritePWSafeFile to save2Path.
	//    So, loadedDb1 and loadedDb2 should be Equal directly.
	equal, err = loadedDb1.Equal(loadedDb2)
	assert.Nil(t, err)
	assert.True(t, equal, "loadedDb1 (after SetRecord and save) and loadedDb2 should be equal")

	// Also check that the original record from simple.dat is still there in loadedDb2
	_, oldRecordExists := loadedDb2.Records["Test entry"]
	assert.True(t, oldRecordExists, "Original 'Test entry' record not found in loadedDb2")
}

func TestAddRecordToPreviouslyEmptyDB(t *testing.T) {
	emptyDb1 := NewV3("", "password")
	emptyDb1.Header.Name = "TestAddAfterEmpty" // As per test description intent
	savePath1 := "./test_dbs/empty_for_add_1.dat"
	emptyDb1.LastSavePath = savePath1 // Set for consistency, though WritePWSafeFile takes path

	// 3. Save emptyDb1
	err := WritePWSafeFile(emptyDb1, savePath1)
	assert.Nil(t, err)
	defer os.Remove(savePath1)

	// 4. Open the saved database into loadedEmptyDb1
	loadedEmptyDb1, err := OpenPWSafeFile(savePath1, "password")
	assert.Nil(t, err)

	// 5. Verify loadedEmptyDb1 is empty
	assert.Equal(t, 0, len(loadedEmptyDb1.Records), "loadedEmptyDb1 should initially be empty")
	assert.Equal(t, "TestAddAfterEmpty", loadedEmptyDb1.Header.Name)

	// 6. Add a new record to loadedEmptyDb1
	var newRec Record
	newRec.Title = "NewRecordInEmpty"
	newRec.Username = "user1"
	newRec.Password = "pass1"
	newRec.Group = "GroupForNew"
	// Other fields will be default zero values, which is fine.
	loadedEmptyDb1.SetRecord(newRec)

	// 7. Assert that loadedEmptyDb1 now contains 1 record
	assert.Equal(t, 1, len(loadedEmptyDb1.Records), "loadedEmptyDb1 should have 1 record after SetRecord")

	// 8. Set path for the next save
	savePath2 := "./test_dbs/empty_then_added_record.dat"
	// loadedEmptyDb1.LastSavePath will be updated by WritePWSafeFile if it's the same path.
	// If different, it's good practice to set it if that field is relied upon by the app,
	// but for WritePWSafeFile itself, the path argument is what matters.
	loadedEmptyDb1.LastSavePath = savePath2 // For clarity and if NeedsSave logic uses it post-save

	// 9. Save loadedEmptyDb1 (which now has one record)
	// This will update loadedEmptyDb1.Header.Version and loadedEmptyDb1.Header.LastSave in memory
	err = WritePWSafeFile(loadedEmptyDb1, savePath2)
	assert.Nil(t, err)
	defer os.Remove(savePath2)

	// 10. Open this second saved database into loadedDbWithRecord
	loadedDbWithRecord, err := OpenPWSafeFile(savePath2, "password")
	assert.Nil(t, err)

	// 11. Verify that loadedDbWithRecord contains the "NewRecordInEmpty" record
	assert.Equal(t, 1, len(loadedDbWithRecord.Records), "loadedDbWithRecord should have 1 record")
	retrievedRec, exists := loadedDbWithRecord.Records["NewRecordInEmpty"]
	assert.True(t, exists, "NewRecordInEmpty should exist in loadedDbWithRecord")
	assert.Equal(t, "user1", retrievedRec.Username)
	assert.Equal(t, "pass1", retrievedRec.Password)
	assert.Equal(t, "GroupForNew", retrievedRec.Group)

	// 13. Use the .Equal() method to compare loadedEmptyDb1 and loadedDbWithRecord
	// After WritePWSafeFile, loadedEmptyDb1's Header.Version and Header.LastSave are updated
	// to match what was written to savePath2. So, loadedDbWithRecord should be Equal to it.
	equal, err := loadedEmptyDb1.Equal(loadedDbWithRecord)
	assert.Nil(t, err)
	assert.True(t, equal, "loadedEmptyDb1 (after SetRecord and save) and loadedDbWithRecord should be equal")
}

// TestNewV3 test creating a new DB, saving it to a file and loading it
func TestNewV3(t *testing.T) {
	newDB := NewV3("", "password")
	var record Record
	record.Title = "Test entry"
	record.Username = "test"
	record.Password = "password"
	record.Group = "test"
	record.URL = "http://test.com"
	record.Notes = "no notes"
	newDB.SetRecord(record)

	newPath := "./test_dbs/simple-new.dat"
	err := WritePWSafeFile(newDB, newPath)
	defer os.Remove(newPath)
	assert.Nil(t, err)

	readNew, err := OpenPWSafeFile("./test_dbs/simple-new.dat", "password")
	assert.Nil(t, err)
	orig, err := OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err)

	// The UUID for these should be different since one was created fresh, check then set the same for comparison
	assert.NotEqual(t, orig.Header.UUID, readNew.Header.UUID)
	readNew.Header.UUID = orig.Header.UUID
	// On write version is set but orig doesn't have it so just set to the same here
	orig.Header.Version = readNew.Header.Version

	equal, err := orig.Equal(readNew)
	assert.Nil(t, err)
	assert.Equal(t, true, equal)
}
