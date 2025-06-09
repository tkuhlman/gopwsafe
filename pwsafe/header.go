package pwsafe

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/pborman/uuid"
	"golang.org/x/crypto/twofish"
)

// header defines the fields in the V3 DB header
// The field is the 1 byte hex value of the field type
type header struct {
	Description    string    `field:"0a"`
	EmptyGroups    []string  `field:"11"`
	Filters        string    `field:"0b"`
	LastSave       time.Time `field:"04"`
	LastSaveBy     []byte    `field:"06"`
	LastSaveHost   []byte    `field:"08"`
	LastSaveUser   []byte    `field:"07"`
	Name           string    `field:"09"`
	PasswordPolicy string    `field:"10"`
	Preferences    string    `field:"02"`
	RecentyUsed    string    `field:"0f"`
	Tree           string    `field:"03"`
	UUID           [16]byte  `field:"01"`
	Version        [2]byte   `field:"00"`
}

func newHeader(name string) header {
	return header{
		Name: name,
		UUID: [16]byte(uuid.NewRandom().Array()),
		// Set the DB version
		Version: [2]byte{0x0E, 0x03},
	}
}

func (h header) Equal(other header) (bool, error) {
	if h.Description != other.Description {
		return false, fmt.Errorf("Description fields not equal, %v != %v", h.Description, other.Description)
	}
	if !slices.Equal(h.EmptyGroups, other.EmptyGroups) {
		return false, fmt.Errorf("EmptyGroups fields not equal, %v != %v", h.EmptyGroups, other.EmptyGroups)
	}
	if h.Filters != other.Filters {
		return false, fmt.Errorf("Filters fields not equal, %v != %v", h.Filters, other.Filters)
	}
	if h.Name != other.Name {
		return false, fmt.Errorf("Name fields not equal, %v != %v", h.Name, other.Name)
	}
	if h.PasswordPolicy != other.PasswordPolicy {
		return false, fmt.Errorf("PasswordPolicy fields not equal, %v != %v", h.PasswordPolicy, other.PasswordPolicy)
	}
	if h.Preferences != other.Preferences {
		return false, fmt.Errorf("Preferences fields not equal, %v != %v", h.Preferences, other.Preferences)
	}
	if h.RecentyUsed != other.RecentyUsed {
		return false, fmt.Errorf("RecentyUsed fields not equal, %v != %v", h.RecentyUsed, other.RecentyUsed)
	}
	if h.Tree != other.Tree {
		return false, fmt.Errorf("Tree fields not equal, %q != %q", h.Tree, other.Tree)
	}
	if h.UUID != other.UUID {
		return false, fmt.Errorf("UUID fields not equal, %v != %v", h.UUID, other.UUID)
	}
	if h.Version != other.Version {
		return false, fmt.Errorf("Version fields not equal, %v != %v", h.Version, other.Version)
	}

	return true, nil
}

// MarshalBinary returns the encoded bytes for the header.
// These are unecrypted and will need encrypting later.
func (h header) MarshalBinary() ([]byte, error) {
	// Note the version field needs to be first and is required, END type field must be last
	//headerBytes, headerValues := marshalRecord(headerFields)
	// TODO the hmac 
	return nil, fmt.Errorf("Not Implemented")
}

// UnmarshalHeader takes a byte slice and unmarshals it into a header struct, also returning the next position in the data and raw bytes
// used so they can be reused for HMAC calculations.
func UnmarshalHeader(data []byte) (header, int, []byte, error) {
	var h header
	headerFieldMap := mapByFieldTag(&h)
	var rdata []byte
	fieldStart := 0
	for {
		if fieldStart > len(data) {
			return h, 0, rdata, errors.New("No END field found when UnMarshaling")
		}
		fieldLength := byteToInt(data[fieldStart : fieldStart+4])
		btype := data[fieldStart+4 : fieldStart+5][0]
		data := data[fieldStart+5 : fieldStart+fieldLength+5]
		rdata = append(rdata, data...)
		fieldStart += fieldLength + 5
		//The next field must start on a block boundary
		blockmod := fieldStart % twofish.BlockSize
		if blockmod != 0 {
			fieldStart += twofish.BlockSize - blockmod
		}

		field, prs := headerFieldMap[btype]
		if prs {
			setField(field, data)
		} else if btype == 0xff { //end
			return h, fieldStart, rdata, nil
		} else {
			return h, fieldStart, rdata, fmt.Errorf("Encountered unknown Record Field type - %v", btype)
		}
	}
}