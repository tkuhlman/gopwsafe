package pwsafe

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

/* The test databases simple.dat and three.dat were made using Loxodo (https://github.com/sommer/loxodo)
Some other test dbs can be found at https://github.com/ronys/pypwsafe/tree/master/test_safes
these all have the password 'bogus12345'
*/

func TestByteToInt(t *testing.T) {
	var testData = []struct {
		bytes []byte
		value int
	}{
		{bytes: []byte{5}, value: 5},
		{bytes: []byte{5, 5}, value: 1285},
		{bytes: []byte{5, 5, 5}, value: 328965},
		{bytes: []byte{5, 5, 5, 5}, value: 84215045},
		{bytes: []byte{255, 255, 255, 255}, value: 4294967295},
	}

	for _, test := range testData {
		derived := byteToInt(test.bytes)
		assert.Equal(t, test.value, derived)
	}
}

func TestCalculateStretchKey(t *testing.T) {
	var db V3
	db.Iter = 2048
	db.Salt = []byte{224, 70, 145, 8, 59, 173, 47, 241, 203, 157, 83, 209, 22, 55, 151, 157, 96, 234, 194, 167, 175, 251, 199, 145, 7, 219, 203, 168, 6, 166, 238, 241}
	expectedKey := [32]byte{243, 201, 143, 194, 139, 58, 186, 186, 133, 14, 238, 200, 139, 153, 45, 247, 215, 251, 24, 49, 28, 170, 157, 181, 21, 174, 129, 231, 234, 62, 51, 203}

	db.calculateStretchKey("password")
	assert.Equal(t, db.StretchedKey, expectedKey)
}

func TestSimpleDB(t *testing.T) {
	// This test relies on the simple password db found at simple.dat
	dbInterface, err := OpenPWSafeFile("./simple.dat", "password")
	assert.Nil(t, err)

	db := dbInterface.(*V3)

	assert.Equal(t, db.GetName(), "")
	assert.Equal(t, len(db.Records), 1)
	record, exists := db.GetRecord("Test entry")
	assert.Equal(t, exists, true)
	assert.Equal(t, record.Username, "test")
	assert.Equal(t, record.Password, "password")
	assert.Equal(t, record.Group, "test")
	assert.Equal(t, record.URL, "http://test.com")
	assert.Equal(t, record.Notes, "no notes")
}

func TestThreeDB(t *testing.T) {
	// This test relies on the password db found at three.dat
	dbInterface, err := OpenPWSafeFile("./three.dat", "three3#;")
	assert.Nil(t, err)

	db := dbInterface.(*V3)

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
	record, exists := db.GetRecord("three entry 1")
	assert.Equal(t, exists, true)
	assert.Equal(t, record.Username, "three1_user")
	assert.Equal(t, record.Password, "three1!@$%^&*()")
	assert.Equal(t, record.Group, "group1")
	assert.Equal(t, record.URL, "http://group1.com")
	assert.Equal(t, record.Notes, "three DB\r\nentry 1")

	//record 2
	record, exists = db.GetRecord("three entry 2")
	assert.Equal(t, exists, true)
	assert.Equal(t, record.Username, "three2_user")
	assert.Equal(t, record.Password, "three2_-+=\\\\|][}{';:")
	assert.Equal(t, record.Group, "group2")
	assert.Equal(t, record.URL, "http://group2.com")
	assert.Equal(t, record.Notes, "three DB\r\nsecond entry")

	//record 3
	record, exists = db.GetRecord("three entry 3")
	assert.Equal(t, exists, true)
	assert.Equal(t, record.Username, "three3_user")
	assert.Equal(t, record.Password, ",./<>?`~0")
	assert.Equal(t, record.Group, "group 3")
	assert.Equal(t, record.URL, "https://group3.com")
	assert.Equal(t, record.Notes, "three DB\r\nentry 3\r\nlast one")
}

func TestInvalidFile(t *testing.T) {
	_, err := OpenPWSafeFile("./db.go", "password")
	assert.Equal(t, err, errors.New("File is not a valid Password Safe v3 file"))
	_, err = OpenPWSafeFile("./notafile", "password")
	assert.NotNil(t, err)
}

func TestBadPassword(t *testing.T) {
	_, err := OpenPWSafeFile("./simple.dat", "badpass")
	assert.Equal(t, err, errors.New("Invalid Password"))
}
