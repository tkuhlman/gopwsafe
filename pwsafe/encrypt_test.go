package pwsafe

import (
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
	err = WritePWSafeFile(&source, "./test_dbs/simple-copy.dat")
	assert.Nil(t, err)
	dest, err := OpenPWSafeFile("./test_dbs/simple-copy.dat", "passwordcopy")
	assert.Nil(t, err)

	equal, err := source.Identical(&dest)
	assert.Nil(t, err)
	assert.Equal(t, true, equal)

	// Reopen the original and verify keys have changed but content is the same
	orig, err := OpenPWSafeFile("./test_dbs/simple.dat", "password")
	assert.Nil(t, err)

	// I expect the stretchedkey, salt, encryption key, hmac key and CBCIV to have changed
	// iter changes also but won't necessarily always
	equal, err = orig.Equal(&dest)
	assert.Nil(t, err)
	assert.Equal(t, true, equal)
	identical, _ := orig.Identical(&dest)
	assert.Equal(t, false, identical)
}
