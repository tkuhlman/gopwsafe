package pwsafe

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/pborman/uuid"
	"golang.org/x/crypto/twofish"
)

// Header field constants
const (
	headerVersion        = 0x00
	headerUUID           = 0x01
	headerPreferences    = 0x02
	headerTree           = 0x03
	headerLastSave       = 0x04
	headerLastSaveBy     = 0x06
	headerLastSaveUser   = 0x07
	headerLastSaveHost   = 0x08
	headerName           = 0x09
	headerDescription    = 0x0a
	headerFilters        = 0x0b
	headerRecentyUsed    = 0x0f
	headerPasswordPolicy = 0x10
	headerEmptyGroups    = 0x11
	headerEndOfEntry     = 0xff
)

// header defines the fields in the V3 DB header
// The field is the 1 byte hex value of the field type
type header struct {
	Description    string    // 0x0a
	EmptyGroups    []string  // 0x11
	Filters        string    // 0x0b
	LastSave       time.Time // 0x04
	LastSaveBy     []byte    // 0x06
	LastSaveHost   []byte    // 0x08
	LastSaveUser   []byte    // 0x07
	Name           string    // 0x09
	PasswordPolicy string    // 0x10
	Preferences    string    // 0x02
	RecentyUsed    string    // 0x0f
	Tree           string    // 0x03
	UUID           [16]byte  // 0x01
	Version        [2]byte   // 0x00
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
		return false, fmt.Errorf("description fields not equal, %v != %v", h.Description, other.Description)
	}
	if !slices.Equal(h.EmptyGroups, other.EmptyGroups) {
		return false, fmt.Errorf("emptyGroups fields not equal, %v != %v", h.EmptyGroups, other.EmptyGroups)
	}
	if h.Filters != other.Filters {
		return false, fmt.Errorf("filters fields not equal, %v != %v", h.Filters, other.Filters)
	}
	if h.Name != other.Name {
		return false, fmt.Errorf("name fields not equal, %v != %v", h.Name, other.Name)
	}
	if h.PasswordPolicy != other.PasswordPolicy {
		return false, fmt.Errorf("passwordPolicy fields not equal, %v != %v", h.PasswordPolicy, other.PasswordPolicy)
	}
	if h.Preferences != other.Preferences {
		return false, fmt.Errorf("preferences fields not equal, %v != %v", h.Preferences, other.Preferences)
	}
	if h.RecentyUsed != other.RecentyUsed {
		return false, fmt.Errorf("recentyUsed fields not equal, %v != %v", h.RecentyUsed, other.RecentyUsed)
	}
	if h.Tree != other.Tree {
		return false, fmt.Errorf("tree fields not equal, %q != %q", h.Tree, other.Tree)
	}
	if h.UUID != other.UUID {
		return false, fmt.Errorf("UUID fields not equal, %v != %v", h.UUID, other.UUID)
	}
	if h.Version != other.Version {
		return false, fmt.Errorf("version fields not equal, %v != %v", h.Version, other.Version)
	}

	return true, nil
}

// setField sets the field value based on the ID
func (h *header) setField(id byte, data []byte) error {
	switch id {
	case headerVersion:
		if len(data) != 2 {
			return fmt.Errorf("invalid length for Version: %d", len(data))
		}
		copy(h.Version[:], data)
	case headerUUID:
		if len(data) != 16 {
			return errors.New("invalid length for UUID")
		}
		copy(h.UUID[:], data)
	case headerPreferences:
		h.Preferences = string(data)
	case headerTree:
		h.Tree = string(data)
	case headerLastSave:
		h.LastSave = time.Unix(int64(binary.LittleEndian.Uint32(data)), 0)
	case headerLastSaveBy:
		h.LastSaveBy = data
	case headerLastSaveUser:
		h.LastSaveUser = data
	case headerLastSaveHost:
		h.LastSaveHost = data
	case headerName:
		h.Name = string(data)
	case headerDescription:
		h.Description = string(data)
	case headerFilters:
		h.Filters = string(data)
	case headerRecentyUsed:
		h.RecentyUsed = string(data)
	case headerPasswordPolicy:
		h.PasswordPolicy = string(data)
	case headerEmptyGroups:
		h.EmptyGroups = append(h.EmptyGroups, string(data))
	default:
		return fmt.Errorf("encountered unknown Header Field type - %v", id)
	}
	return nil
}

// marshal returns the binary format for the header and the values used for hmac calculations
func (h *header) marshal() ([]byte, []byte) {
	var recordBuf bytes.Buffer
	var hmacBuf bytes.Buffer

	// Helper to append a field
	appendField := func(id byte, data any) {
		size := binary.Size(data)
		if size <= 0 {
			return
		}
		// Write to HMAC buffer
		binary.Write(&hmacBuf, binary.LittleEndian, data)

		// Write length
		binary.Write(&recordBuf, binary.LittleEndian, uint32(size))
		// Write ID
		recordBuf.WriteByte(id)
		// Write Data
		binary.Write(&recordBuf, binary.LittleEndian, data)

		// Write Padding
		usedBlockSpace := (size + 5) % twofish.BlockSize
		if usedBlockSpace != 0 {
			recordBuf.Write(pseudoRandomBytes(twofish.BlockSize - usedBlockSpace))
		}
	}

	// Version is required and should be first
	appendField(headerVersion, h.Version[:])
	appendField(headerUUID, h.UUID[:])
	appendField(headerPreferences, []byte(h.Preferences))
	appendField(headerTree, []byte(h.Tree))
	if !h.LastSave.IsZero() {
		appendField(headerLastSave, uint32(h.LastSave.Unix()))
	}
	appendField(headerLastSaveBy, h.LastSaveBy)
	appendField(headerLastSaveUser, h.LastSaveUser)
	appendField(headerLastSaveHost, h.LastSaveHost)
	appendField(headerName, []byte(h.Name))
	appendField(headerDescription, []byte(h.Description))
	appendField(headerFilters, []byte(h.Filters))
	appendField(headerRecentyUsed, []byte(h.RecentyUsed))
	appendField(headerPasswordPolicy, []byte(h.PasswordPolicy))
	for _, group := range h.EmptyGroups {
		appendField(headerEmptyGroups, []byte(group))
	}

	// End of entry
	recordBuf.Write([]byte{0, 0, 0, 0})
	recordBuf.WriteByte(headerEndOfEntry)
	recordBuf.Write(pseudoRandomBytes(twofish.BlockSize - 5))

	return recordBuf.Bytes(), hmacBuf.Bytes()
}

// UnmarshalHeader takes a byte slice and unmarshals it into a header struct, also returning the next position in the data and raw bytes
// used so they can be reused for HMAC calculations.
func UnmarshalHeader(data []byte) (header, int, []byte, error) {
	var h header
	var rdata []byte
	fieldStart := 0
	for {
		if fieldStart+5 > len(data) {
			return h, 0, rdata, errors.New("no END field found when UnMarshaling")
		}
		fieldLength := int(binary.LittleEndian.Uint32(data[fieldStart : fieldStart+4]))
		btype := data[fieldStart+4 : fieldStart+5][0]
		if fieldStart+fieldLength+5 > len(data) {
			return h, fieldStart, rdata, fmt.Errorf("invalid field length %d at offset %d, exceeds data length %d", fieldLength, fieldStart, len(data))
		}
		fieldData := data[fieldStart+5 : fieldStart+fieldLength+5]
		rdata = append(rdata, fieldData...)
		fieldStart += fieldLength + 5
		//The next field must start on a block boundary
		blockmod := fieldStart % twofish.BlockSize
		if blockmod != 0 {
			fieldStart += twofish.BlockSize - blockmod
		}

		if btype == headerEndOfEntry {
			return h, fieldStart, rdata, nil
		}

		if err := h.setField(btype, fieldData); err != nil {
			// For forward compatibility, maybe we should ignore unknown fields?
			// But the original code returned error.
			return h, fieldStart, rdata, err
		}
	}
}
