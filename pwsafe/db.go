// The database type for a Password Safe V3 database
// The db specification - http://sourceforge.net/p/passwordsafe/code/HEAD/tree/trunk/pwsafe/pwsafe/docs/formatV3.txt

package pwsafe

import (
	"time"
//	"code.google.com/p/go.crypto/twofish"
	"uuid"
)

type Record struct {
	AccessTime	time.Time
	CreateTime	time.Time
	Group				string
	ModTime			time.Time
	Notes				string
	Password		string
	PasswordModTime	string
	Title				string
	Username		string
	URL					string
	UUID				uuid.UUID

}

type DB struct {
	// Note not all of the Header information from the specification is implemented
	Name        string
	Description string
	LastSave    time.Time
	Records     map[string]Record //the key is the record title
	UUID        uuid.UUID
	Version     string
}
