package pwsafe

import (
	"fmt"
	"time"
)

// Record The primary type for password DB entries
type Record struct {
	AccessTime             time.Time `field:"09"`
	Autotype               string    `field:"0e"`
	CreateTime             time.Time `field:"07"`
	DoubleClickAction      [2]byte   `field:"13"`
	Email                  string    `field:"14"`
	Group                  string    `field:"02"`
	ModTime                time.Time `field:"0c"`
	Notes                  string    `field:"05"`
	Password               string    `field:"06"`
	PasswordExpiry         time.Time `field:"0a"`
	PasswordExpiryInterval [4]byte   `field:"11"`
	PasswordHistory        string    `field:"0f"`
	PasswordModTime        string    `field:"08"`
	PasswordPolicy         string    `field:"10"`
	PasswordPolicyName     string    `field:"18"`
	ProtectedEntry         byte      `field:"15"`
	RunCommand             string    `field:"12"`
	ShiftDoubleClickAction [2]byte   `field:"17"`
	Title                  string    `field:"03"`
	Username               string    `field:"04"`
	URL                    string    `field:"0d"`
	UUID                   [16]byte  `field:"01"`
}

// Equal compares two records returning true optionally skipping create/access times and UUID.
func (r Record) Equal(otherRecord Record, skipTimes bool) (bool, error) {
	if r.Autotype != otherRecord.Autotype {
		return false, fmt.Errorf("Records don't match, Autotype: %v != %v", r.Autotype, otherRecord.Autotype)
	}
	if r.DoubleClickAction != otherRecord.DoubleClickAction {
		return false, fmt.Errorf("Records don't match, DoubleClickAction: %v != %v", r.DoubleClickAction, otherRecord.DoubleClickAction)
	}
	if r.Email != otherRecord.Email {
		return false, fmt.Errorf("Records don't match, Email: %v != %v", r.Email, otherRecord.Email)
	}
	if r.Group != otherRecord.Group {
		return false, fmt.Errorf("Records don't match, Group: %v != %v", r.Group, otherRecord.Group)
	}
	if r.Notes != otherRecord.Notes {
		return false, fmt.Errorf("Records don't match, Notes: %v != %v", r.Notes, otherRecord.Notes)
	}
	if r.Password != otherRecord.Password {
		return false, fmt.Errorf("Records don't match, Password: %v != %v", r.Password, otherRecord.Password)
	}
	if !r.PasswordExpiry.Equal(otherRecord.PasswordExpiry) {
		return false, fmt.Errorf("Records don't match, PasswordExpiry: %v != %v", r.PasswordExpiry, otherRecord.PasswordExpiry)
	}
	if r.PasswordExpiryInterval != otherRecord.PasswordExpiryInterval {
		return false, fmt.Errorf("Records don't match, PasswordExpiryInterval: %v != %v", r.PasswordExpiryInterval, otherRecord.PasswordExpiryInterval)
	}
	if r.PasswordHistory != otherRecord.PasswordHistory {
		return false, fmt.Errorf("Records don't match, PasswordHistory: %v != %v", r.PasswordHistory, otherRecord.PasswordHistory)
	}
	if r.PasswordModTime != otherRecord.PasswordModTime {
		return false, fmt.Errorf("Records don't match, PasswordModTime: %v != %v", r.PasswordModTime, otherRecord.PasswordModTime)
	}
	if r.PasswordPolicy != otherRecord.PasswordPolicy {
		return false, fmt.Errorf("Records don't match, PasswordPolicy: %v != %v", r.PasswordPolicy, otherRecord.PasswordPolicy)
	}
	if r.PasswordPolicyName != otherRecord.PasswordPolicyName {
		return false, fmt.Errorf("Records don't match, PasswordPolicyName: %v != %v", r.PasswordPolicyName, otherRecord.PasswordPolicyName)
	}
	if r.ProtectedEntry != otherRecord.ProtectedEntry {
		return false, fmt.Errorf("Records don't match, ProtectedEntry: %v != %v", r.ProtectedEntry, otherRecord.ProtectedEntry)
	}
	if r.RunCommand != otherRecord.RunCommand {
		return false, fmt.Errorf("Records don't match, RunCommand: %v != %v", r.RunCommand, otherRecord.RunCommand)
	}
	if r.ShiftDoubleClickAction != otherRecord.ShiftDoubleClickAction {
		return false, fmt.Errorf("Records don't match, ShiftDoubleClickAction: %v != %v", r.ShiftDoubleClickAction, otherRecord.ShiftDoubleClickAction)
	}
	if r.Title != otherRecord.Title {
		return false, fmt.Errorf("Records don't match, Title: %v != %v", r.Title, otherRecord.Title)
	}
	if r.Username != otherRecord.Username {
		return false, fmt.Errorf("Records don't match, Username: %v != %v", r.Username, otherRecord.Username)
	}
	if r.URL != otherRecord.URL {
		return false, fmt.Errorf("Records don't match, URL: %v != %v", r.URL, otherRecord.URL)
	}

	if !skipTimes {
		if !r.AccessTime.Equal(otherRecord.AccessTime) {
			return false, fmt.Errorf("Records don't match, AccessTime: %v != %v", r.AccessTime, otherRecord.AccessTime)
		}
		if !r.CreateTime.Equal(otherRecord.CreateTime) {
			return false, fmt.Errorf("Records don't match, CreateTime: %v != %v", r.CreateTime, otherRecord.CreateTime)
		}
		if !r.ModTime.Equal(otherRecord.ModTime) {
			return false, fmt.Errorf("Records don't match, ModTime: %v != %v", r.ModTime, otherRecord.ModTime)
		}
		if r.UUID != otherRecord.UUID {
			return false, fmt.Errorf("Records don't match, UUID: %v != %v", r.UUID, otherRecord.UUID)
		}
	}

	return true, nil
}
