package pwsafe

import (
	"fmt"
	"slices"
)

// needsSave returns true if the db has unsaved modifications.
func needsSave(db *V3) bool {
	return db.Header.LastSave.Before(db.LastMod)
}

// recordByTitle finds a record by title via linear scan.
// Only for tests loading pre-existing .dat files where the UUID is not known in advance.
func recordByTitle(db *V3, title string) (Record, bool) {
	for _, rec := range db.Records {
		if rec.Title == title {
			return rec, true
		}
	}
	return Record{}, false
}

// deleteByTitle deletes a record by title. Returns true if found and deleted.
// Only for tests loading pre-existing .dat files where the UUID is not known in advance.
func deleteByTitle(db *V3, title string) bool {
	for k, rec := range db.Records {
		if rec.Title == title {
			db.DeleteRecord(k)
			return true
		}
	}
	return false
}

// dbEqual compares two V3 databases for content equality (test use only).
func dbEqual(a, b *V3) (bool, error) {
	if ok, err := headerEqual(a.Header, b.Header); !ok {
		return false, err
	}
	if len(a.Records) != len(b.Records) {
		return false, fmt.Errorf("record count mismatch: %d != %d", len(a.Records), len(b.Records))
	}
	for uuidHex, rec := range a.Records {
		other, ok := b.Records[uuidHex]
		if !ok {
			return false, fmt.Errorf("record %q not found in other DB", rec.Title)
		}
		if ok, err := recordEqual(rec, other); !ok {
			return false, err
		}
	}
	return true, nil
}

func headerEqual(h, other header) (bool, error) {
	if h.Description != other.Description {
		return false, fmt.Errorf("description: %q != %q", h.Description, other.Description)
	}
	if !slices.Equal(h.EmptyGroups, other.EmptyGroups) {
		return false, fmt.Errorf("emptyGroups: %v != %v", h.EmptyGroups, other.EmptyGroups)
	}
	if h.Filters != other.Filters {
		return false, fmt.Errorf("filters: %v != %v", h.Filters, other.Filters)
	}
	if h.Name != other.Name {
		return false, fmt.Errorf("name: %q != %q", h.Name, other.Name)
	}
	if h.PasswordPolicy != other.PasswordPolicy {
		return false, fmt.Errorf("passwordPolicy: %v != %v", h.PasswordPolicy, other.PasswordPolicy)
	}
	if h.Preferences != other.Preferences {
		return false, fmt.Errorf("preferences: %v != %v", h.Preferences, other.Preferences)
	}
	if h.RecentyUsed != other.RecentyUsed {
		return false, fmt.Errorf("recentyUsed: %v != %v", h.RecentyUsed, other.RecentyUsed)
	}
	if h.Tree != other.Tree {
		return false, fmt.Errorf("tree: %q != %q", h.Tree, other.Tree)
	}
	if h.UUID != other.UUID {
		return false, fmt.Errorf("UUID: %v != %v", h.UUID, other.UUID)
	}
	if h.Version != other.Version {
		return false, fmt.Errorf("version: %v != %v", h.Version, other.Version)
	}
	return true, nil
}

func recordEqual(r, other Record) (bool, error) {
	if r.Autotype != other.Autotype {
		return false, fmt.Errorf("Autotype: %v != %v", r.Autotype, other.Autotype)
	}
	if r.DoubleClickAction != other.DoubleClickAction {
		return false, fmt.Errorf("DoubleClickAction: %v != %v", r.DoubleClickAction, other.DoubleClickAction)
	}
	if r.Email != other.Email {
		return false, fmt.Errorf("Email: %v != %v", r.Email, other.Email)
	}
	if r.Group != other.Group {
		return false, fmt.Errorf("Group: %v != %v", r.Group, other.Group)
	}
	if r.Notes != other.Notes {
		return false, fmt.Errorf("Notes: %v != %v", r.Notes, other.Notes)
	}
	if r.OwnSymbolsForPassword != other.OwnSymbolsForPassword {
		return false, fmt.Errorf("OwnSymbolsForPassword: %v != %v", r.OwnSymbolsForPassword, other.OwnSymbolsForPassword)
	}
	if r.Password != other.Password {
		return false, fmt.Errorf("Password mismatch")
	}
	if !r.PasswordExpiry.Equal(other.PasswordExpiry) {
		return false, fmt.Errorf("PasswordExpiry: %v != %v", r.PasswordExpiry, other.PasswordExpiry)
	}
	if r.PasswordExpiryInterval != other.PasswordExpiryInterval {
		return false, fmt.Errorf("PasswordExpiryInterval: %v != %v", r.PasswordExpiryInterval, other.PasswordExpiryInterval)
	}
	if r.PasswordHistory != other.PasswordHistory {
		return false, fmt.Errorf("PasswordHistory: %v != %v", r.PasswordHistory, other.PasswordHistory)
	}
	if r.PasswordModTime != other.PasswordModTime {
		return false, fmt.Errorf("PasswordModTime: %v != %v", r.PasswordModTime, other.PasswordModTime)
	}
	if r.PasswordPolicy != other.PasswordPolicy {
		return false, fmt.Errorf("PasswordPolicy: %v != %v", r.PasswordPolicy, other.PasswordPolicy)
	}
	if r.PasswordPolicyName != other.PasswordPolicyName {
		return false, fmt.Errorf("PasswordPolicyName: %v != %v", r.PasswordPolicyName, other.PasswordPolicyName)
	}
	if r.ProtectedEntry != other.ProtectedEntry {
		return false, fmt.Errorf("ProtectedEntry: %v != %v", r.ProtectedEntry, other.ProtectedEntry)
	}
	if r.RunCommand != other.RunCommand {
		return false, fmt.Errorf("RunCommand: %v != %v", r.RunCommand, other.RunCommand)
	}
	if r.ShiftDoubleClickAction != other.ShiftDoubleClickAction {
		return false, fmt.Errorf("ShiftDoubleClickAction: %v != %v", r.ShiftDoubleClickAction, other.ShiftDoubleClickAction)
	}
	if r.Title != other.Title {
		return false, fmt.Errorf("Title: %q != %q", r.Title, other.Title)
	}
	if r.URL != other.URL {
		return false, fmt.Errorf("URL: %v != %v", r.URL, other.URL)
	}
	if r.Username != other.Username {
		return false, fmt.Errorf("Username: %v != %v", r.Username, other.Username)
	}
	return true, nil
}
