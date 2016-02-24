package pwsafe

import "github.com/fatih/structs"

// Equal returns true if the two dbs have the same data but not necessarily the same keys
func (db *V3) Equal(other *DB) bool {
	// restrict comparison to fields with a field struct tag
	otherStruct := structs.New(other)
	for _, field := range mapByFieldTag(db) {
		if field.Value() != otherStruct.Field(field.Name()).Value() {
			return false
		}
	}

	// compare records
	if len(db.List()) != len((*other).List()) {
		return false
	}
	// todo start an interface for Record and add an Equal method, direct comparison fails because of the UUID, alternatively store UUID as a string
	//	for _, title := range db.List() {
	//		dbRecord, _ := db.GetRecord(title)
	//		otherRecord, _ := (*other).GetRecord(title)
	//		if dbRecord != otherRecord {
	//			return false
	//		}
	//	}
	return true
}

// Identical returns true if the two dbs have the same fields including the cryptographic keys
func (db *V3) Identical(other *DB) bool {
	if db.Equal(other) == false {
		return false
	}
	dbStruct := structs.New(db)
	otherStruct := structs.New(other)
	for _, fieldName := range []string{"CBCIV", "encryptionKey", "HMACKEY", "Iter", "Salt", "stretchedKey"} {
		if dbStruct.Field(fieldName).Value() != otherStruct.Field(fieldName).Value() {
			return false
		}
	}

	return true
}
