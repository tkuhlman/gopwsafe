package pwsafe

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/twofish"
)

// Record field constants
const (
	recordUUID                   = 0x01
	recordGroup                  = 0x02
	recordTitle                  = 0x03
	recordUsername               = 0x04
	recordNotes                  = 0x05
	recordPassword               = 0x06
	recordCreateTime             = 0x07
	recordPasswordModTime        = 0x08
	recordAccessTime             = 0x09
	recordPasswordExpiry         = 0x0a
	recordModTime                = 0x0c
	recordURL                    = 0x0d
	recordAutotype               = 0x0e
	recordPasswordHistory        = 0x0f
	recordPasswordPolicy         = 0x10
	recordPasswordExpiryInterval = 0x11
	recordRunCommand             = 0x12
	recordDoubleClickAction      = 0x13
	recordEmail                  = 0x14
	recordProtectedEntry         = 0x15
	recordShiftDoubleClickAction = 0x17
	recordPasswordPolicyName     = 0x18
	recordEndOfEntry             = 0xff
)

// Record The primary type for password DB entries
type Record struct {
	AccessTime             time.Time // 0x09
	Autotype               string    // 0x0e
	CreateTime             time.Time // 0x07
	DoubleClickAction      [2]byte   // 0x13
	Email                  string    // 0x14
	Group                  string    // 0x02
	ModTime                time.Time // 0x0c
	Notes                  string    // 0x05
	Password               string    // 0x06
	PasswordExpiry         time.Time // 0x0a
	PasswordExpiryInterval [4]byte   // 0x11
	PasswordHistory        string    // 0x0f
	PasswordModTime        string    // 0x08
	PasswordPolicy         string    // 0x10
	PasswordPolicyName     string    // 0x18
	ProtectedEntry         byte      // 0x15
	RunCommand             string    // 0x12
	ShiftDoubleClickAction [2]byte   // 0x17
	Title                  string    // 0x03
	Username               string    // 0x04
	URL                    string    // 0x0d
	UUID                   [16]byte  // 0x01
}

// Equal compares two records returning true optionally skipping create/access times and UUID.
func (r Record) Equal(otherRecord Record, skipTimes bool) (bool, error) {
	if r.Autotype != otherRecord.Autotype {
		return false, fmt.Errorf("records don't match, Autotype: %v != %v", r.Autotype, otherRecord.Autotype)
	}
	if r.DoubleClickAction != otherRecord.DoubleClickAction {
		return false, fmt.Errorf("records don't match, DoubleClickAction: %v != %v", r.DoubleClickAction, otherRecord.DoubleClickAction)
	}
	if r.Email != otherRecord.Email {
		return false, fmt.Errorf("records don't match, Email: %v != %v", r.Email, otherRecord.Email)
	}
	if r.Group != otherRecord.Group {
		return false, fmt.Errorf("records don't match, Group: %v != %v", r.Group, otherRecord.Group)
	}
	if r.Notes != otherRecord.Notes {
		return false, fmt.Errorf("records don't match, Notes: %v != %v", r.Notes, otherRecord.Notes)
	}
	if r.Password != otherRecord.Password {
		return false, fmt.Errorf("records don't match, Password: %v != %v", r.Password, otherRecord.Password)
	}
	if !r.PasswordExpiry.Equal(otherRecord.PasswordExpiry) {
		return false, fmt.Errorf("records don't match, PasswordExpiry: %v != %v", r.PasswordExpiry, otherRecord.PasswordExpiry)
	}
	if r.PasswordExpiryInterval != otherRecord.PasswordExpiryInterval {
		return false, fmt.Errorf("records don't match, PasswordExpiryInterval: %v != %v", r.PasswordExpiryInterval, otherRecord.PasswordExpiryInterval)
	}
	if r.PasswordHistory != otherRecord.PasswordHistory {
		return false, fmt.Errorf("records don't match, PasswordHistory: %v != %v", r.PasswordHistory, otherRecord.PasswordHistory)
	}
	if r.PasswordModTime != otherRecord.PasswordModTime {
		return false, fmt.Errorf("records don't match, PasswordModTime: %v != %v", r.PasswordModTime, otherRecord.PasswordModTime)
	}
	if r.PasswordPolicy != otherRecord.PasswordPolicy {
		return false, fmt.Errorf("records don't match, PasswordPolicy: %v != %v", r.PasswordPolicy, otherRecord.PasswordPolicy)
	}
	if r.PasswordPolicyName != otherRecord.PasswordPolicyName {
		return false, fmt.Errorf("records don't match, PasswordPolicyName: %v != %v", r.PasswordPolicyName, otherRecord.PasswordPolicyName)
	}
	if r.ProtectedEntry != otherRecord.ProtectedEntry {
		return false, fmt.Errorf("records don't match, ProtectedEntry: %v != %v", r.ProtectedEntry, otherRecord.ProtectedEntry)
	}
	if r.RunCommand != otherRecord.RunCommand {
		return false, fmt.Errorf("records don't match, RunCommand: %v != %v", r.RunCommand, otherRecord.RunCommand)
	}
	if r.ShiftDoubleClickAction != otherRecord.ShiftDoubleClickAction {
		return false, fmt.Errorf("records don't match, ShiftDoubleClickAction: %v != %v", r.ShiftDoubleClickAction, otherRecord.ShiftDoubleClickAction)
	}
	if r.Title != otherRecord.Title {
		return false, fmt.Errorf("records don't match, Title: %v != %v", r.Title, otherRecord.Title)
	}
	if r.Username != otherRecord.Username {
		return false, fmt.Errorf("records don't match, Username: %v != %v", r.Username, otherRecord.Username)
	}
	if r.URL != otherRecord.URL {
		return false, fmt.Errorf("records don't match, URL: %v != %v", r.URL, otherRecord.URL)
	}

	if !skipTimes {
		if !r.AccessTime.Equal(otherRecord.AccessTime) {
			return false, fmt.Errorf("records don't match, AccessTime: %v != %v", r.AccessTime, otherRecord.AccessTime)
		}
		if !r.CreateTime.Equal(otherRecord.CreateTime) {
			return false, fmt.Errorf("records don't match, CreateTime: %v != %v", r.CreateTime, otherRecord.CreateTime)
		}
		if !r.ModTime.Equal(otherRecord.ModTime) {
			return false, fmt.Errorf("records don't match, ModTime: %v != %v", r.ModTime, otherRecord.ModTime)
		}
		if r.UUID != otherRecord.UUID {
			return false, fmt.Errorf("Records don't match, UUID: %v != %v", r.UUID, otherRecord.UUID)
		}
	}

	return true, nil
}

// setField sets the field value based on the ID
func (r *Record) setField(id byte, data []byte) error {
	switch id {
	case recordUUID:
		if len(data) != 16 {
			return errors.New("invalid length for UUID")
		}
		copy(r.UUID[:], data)
	case recordGroup:
		r.Group = string(data)
	case recordTitle:
		r.Title = string(data)
	case recordUsername:
		r.Username = string(data)
	case recordNotes:
		r.Notes = string(data)
	case recordPassword:
		r.Password = string(data)
	case recordCreateTime:
		r.CreateTime = time.Unix(int64(binary.LittleEndian.Uint32(data)), 0)
	case recordPasswordModTime:
		r.PasswordModTime = string(data)
	case recordAccessTime:
		r.AccessTime = time.Unix(int64(binary.LittleEndian.Uint32(data)), 0)
	case recordPasswordExpiry:
		r.PasswordExpiry = time.Unix(int64(binary.LittleEndian.Uint32(data)), 0)
	case recordModTime:
		r.ModTime = time.Unix(int64(binary.LittleEndian.Uint32(data)), 0)
	case recordURL:
		r.URL = string(data)
	case recordAutotype:
		r.Autotype = string(data)
	case recordPasswordHistory:
		r.PasswordHistory = string(data)
	case recordPasswordPolicy:
		r.PasswordPolicy = string(data)
	case recordPasswordExpiryInterval:
		if len(data) != 4 {
			return errors.New("invalid length for PasswordExpiryInterval")
		}
		copy(r.PasswordExpiryInterval[:], data)
	case recordRunCommand:
		r.RunCommand = string(data)
	case recordDoubleClickAction:
		if len(data) != 2 {
			return errors.New("invalid length for DoubleClickAction")
		}
		copy(r.DoubleClickAction[:], data)
	case recordEmail:
		r.Email = string(data)
	case recordProtectedEntry:
		if len(data) != 1 {
			return errors.New("invalid length for ProtectedEntry")
		}
		r.ProtectedEntry = data[0]
	case recordShiftDoubleClickAction:
		if len(data) != 2 {
			return errors.New("invalid length for ShiftDoubleClickAction")
		}
		copy(r.ShiftDoubleClickAction[:], data)
	case recordPasswordPolicyName:
		r.PasswordPolicyName = string(data)
	default:
		return fmt.Errorf("encountered unknown Record Field type - %v", id)
	}
	return nil
}

// marshal returns the binary format for the record and the values used for hmac calculations
func (r *Record) marshal() ([]byte, []byte) {
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
			recordBuf.Write(pseudoRandmonBytes(twofish.BlockSize - usedBlockSpace))
		}
	}

	appendField(recordUUID, r.UUID[:])
	appendField(recordGroup, []byte(r.Group))
	appendField(recordTitle, []byte(r.Title))
	appendField(recordUsername, []byte(r.Username))
	appendField(recordNotes, []byte(r.Notes))
	appendField(recordPassword, []byte(r.Password))
	if !r.CreateTime.IsZero() {
		appendField(recordCreateTime, uint32(r.CreateTime.Unix()))
	}
	appendField(recordPasswordModTime, []byte(r.PasswordModTime))
	if !r.AccessTime.IsZero() {
		appendField(recordAccessTime, uint32(r.AccessTime.Unix()))
	}
	if !r.PasswordExpiry.IsZero() {
		appendField(recordPasswordExpiry, uint32(r.PasswordExpiry.Unix()))
	}
	if !r.ModTime.IsZero() {
		appendField(recordModTime, uint32(r.ModTime.Unix()))
	}
	appendField(recordURL, []byte(r.URL))
	appendField(recordAutotype, []byte(r.Autotype))
	appendField(recordPasswordHistory, []byte(r.PasswordHistory))
	appendField(recordPasswordPolicy, []byte(r.PasswordPolicy))
	appendField(recordPasswordExpiryInterval, r.PasswordExpiryInterval[:])
	appendField(recordRunCommand, []byte(r.RunCommand))
	appendField(recordDoubleClickAction, r.DoubleClickAction[:])
	appendField(recordEmail, []byte(r.Email))
	if r.ProtectedEntry != 0 {
		appendField(recordProtectedEntry, []byte{r.ProtectedEntry})
	}
	appendField(recordShiftDoubleClickAction, r.ShiftDoubleClickAction[:])
	appendField(recordPasswordPolicyName, []byte(r.PasswordPolicyName))

	// End of entry
	recordBuf.Write([]byte{0, 0, 0, 0})
	recordBuf.WriteByte(recordEndOfEntry)
	recordBuf.Write(pseudoRandmonBytes(twofish.BlockSize - 5))

	return recordBuf.Bytes(), hmacBuf.Bytes()
}
