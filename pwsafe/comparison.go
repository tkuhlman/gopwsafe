package pwsafe

import (
	"fmt"
	"reflect"

	"github.com/fatih/structs"
)

// Equal returns true if the two dbs have the same data but not necessarily the same keys nor same LastSave time
func (db *V3) Equal(other *DB) (bool, error) {
	// restrict comparison to fields with a field struct tag
	otherStruct := structs.New(*other)
	for _, field := range mapByFieldTag(db) {
		// todo should I compare version?
		if field.Name() == "LastSave" || field.Name() == "Version" {
			continue
		}
		if !reflect.DeepEqual(field.Value(), otherStruct.Field(field.Name()).Value()) {
			return false, fmt.Errorf("%v fields not equal, %v != %v", field.Name(), field.Value(), otherStruct.Field(field.Name()).Value())
		}
	}

	// compare records
	if len(db.List()) != len((*other).List()) {
		return false, fmt.Errorf("record lengths don't match, %v != %v", len(db.List()), len((*other).List()))
	}
	for _, title := range db.List() {
		dbRecord, _ := db.GetRecord(title)
		otherRecord, _ := (*other).GetRecord(title)
		if !reflect.DeepEqual(dbRecord, otherRecord) {
			return false, fmt.Errorf("Records don't match, %v != %v", dbRecord, otherRecord)
		}
	}
	return true, nil
}

// Identical returns true if the two dbs have the same fields including the cryptographic keys
func (db *V3) Identical(other *DB) (bool, error) {
	equal, err := db.Equal(other)
	if !equal {
		return false, err
	}
	dbStruct := structs.New(*db)
	otherStruct := structs.New(*other)
	for _, fieldName := range []string{"CBCIV", "EncryptionKey", "HMACKey", "Iter", "Salt", "StretchedKey"} {
		if !reflect.DeepEqual(dbStruct.Field(fieldName).Value(), otherStruct.Field(fieldName).Value()) {
			return false, fmt.Errorf("%v fields not equal, %v != %v", fieldName, dbStruct.Field(fieldName).Value(), otherStruct.Field(fieldName).Value())
		}
	}

	return true, nil
}
