package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/tkuhlman/gopwsafe/pwsafe"
)

var db *pwsafe.V3

func openDB(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return "invalid arguments: expected (data, password)"
	}

	dataJS := args[0]
	password := args[1].String()

	// Copy data from JS Uint8Array to Go []byte
	length := dataJS.Get("length").Int()
	data := make([]byte, length)
	js.CopyBytesToGo(data, dataJS)

	reader := bytes.NewReader(data)

	newDB := &pwsafe.V3{}
	_, err := newDB.Decrypt(reader, password)
	if err != nil {
		return fmt.Sprintf("failed to decrypt: %s", err)
	}

	db = newDB
	return nil // null means success
}

func getDBData(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}

	// We want to return a tree structure.
	// For now, let's just return a flat list of items with their groups for the frontend to process,
	// or build a simplified tree here.
	// Let's return a list of {uuid, title, group} objects.

	type Item struct {
		UUID  string `json:"uuid"`
		Title string `json:"title"`
		Group string `json:"group"`
	}

	var items []Item

	keys := db.List()
	for _, title := range keys {
		rec := db.Records[title]
		// UUID is [16]byte, need to convert to string
		uuidStr := fmt.Sprintf("%x", rec.UUID)
		items = append(items, Item{
			UUID:  uuidStr,
			Title: rec.Title,
			Group: rec.Group,
		})
	}

	jsonData, err := json.Marshal(items)
	if err != nil {
		return fmt.Sprintf("json marshal error: %s", err)
	}

	return string(jsonData)
}

func getRecord(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}
	if len(args) != 1 {
		return "invalid arguments: expected (title)" // Using title as key for now based on map
	}

	title := args[0].String()
	rec, ok := db.Records[title]
	if !ok {
		return "record not found"
	}

	// We shouldn't return the raw struct if it contains sensitive binary data that doesn't JSON marshal well slightly,
	// but Record struct has tags? Let's check record.go later.
	// For now assuming json.Marshal works or we create a DTO.

	jsonData, err := json.Marshal(rec)
	if err != nil {
		return fmt.Sprintf("json marshal error: %s", err)
	}

	return string(jsonData)
}

func createDatabase(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return "invalid arguments: expected (password)"
	}
	password := args[0].String()

	newDB := pwsafe.NewV3("", password)
	db = newDB
	return nil
}

func getDBInfo(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}
	// Return header info
	// db.Header contains the info.
	type DBInfo struct {
		Version     string `json:"version"`
		UUID        string `json:"uuid"`
		Description string `json:"description"`
		What        string `json:"what"`
		When        string `json:"when"`
		Who         string `json:"who"`
	}

	// UUID to string
	uuidStr := fmt.Sprintf("%x", db.Header.UUID)

	info := DBInfo{
		Version:     fmt.Sprintf("%x", db.Header.Version),
		UUID:        uuidStr,
		Description: db.Header.Description,
		What:        string(db.Header.LastSaveBy),
		When:        db.Header.LastSave.String(),
		Who:         string(db.Header.LastSaveUser),
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		return fmt.Sprintf("json marshal error: %s", err)
	}

	return string(jsonData)
}

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("openDB", js.FuncOf(openDB))
	js.Global().Set("getDBData", js.FuncOf(getDBData))
	js.Global().Set("getRecord", js.FuncOf(getRecord))
	js.Global().Set("createDatabase", js.FuncOf(createDatabase))
	js.Global().Set("getDBInfo", js.FuncOf(getDBInfo))
	js.Global().Set("saveDB", js.FuncOf(saveDB))
	js.Global().Set("addRecord", js.FuncOf(addRecord))
	js.Global().Set("updateRecord", js.FuncOf(updateRecord))
	js.Global().Set("deleteRecord", js.FuncOf(deleteRecord))

	fmt.Println("WASM initialized")
	<-c
}

// RecordDTO is a data transfer object for Record to/from JSON
type RecordDTO struct {
	AccessTime             string   `json:"accessTime"`
	Autotype               string   `json:"autotype"`
	CreateTime             string   `json:"createTime"`
	DoubleClickAction      [2]byte  `json:"doubleClickAction"`
	Email                  string   `json:"email"`
	Group                  string   `json:"group"`
	ModTime                string   `json:"modTime"`
	Notes                  string   `json:"notes"`
	Password               string   `json:"password"`
	PasswordExpiry         string   `json:"passwordExpiry"`
	PasswordExpiryInterval uint32   `json:"passwordExpiryInterval"`
	PasswordHistory        string   `json:"passwordHistory"`
	PasswordModTime        string   `json:"passwordModTime"`
	PasswordPolicy         string   `json:"passwordPolicy"`
	PasswordPolicyName     string   `json:"passwordPolicyName"`
	ProtectedEntry         byte     `json:"protectedEntry"`
	RunCommand             string   `json:"runCommand"`
	ShiftDoubleClickAction [2]byte  `json:"shiftDoubleClickAction"`
	Title                  string   `json:"title"`
	Username               string   `json:"username"`
	URL                    string   `json:"url"`
	UUID                   [16]byte `json:"uuid"`
}

func (dto *RecordDTO) toRecord() (pwsafe.Record, error) {
	var r pwsafe.Record
	r.AccessTime, _ = time.Parse(time.RFC3339, dto.AccessTime)
	r.Autotype = dto.Autotype
	r.CreateTime, _ = time.Parse(time.RFC3339, dto.CreateTime)
	r.DoubleClickAction = dto.DoubleClickAction
	r.Email = dto.Email
	r.Group = dto.Group
	r.ModTime, _ = time.Parse(time.RFC3339, dto.ModTime)
	r.Notes = dto.Notes
	r.Password = dto.Password
	r.PasswordExpiry, _ = time.Parse(time.RFC3339, dto.PasswordExpiry)
	r.PasswordExpiryInterval = dto.PasswordExpiryInterval
	r.PasswordHistory = dto.PasswordHistory
	r.PasswordModTime = dto.PasswordModTime
	r.PasswordPolicy = dto.PasswordPolicy
	r.PasswordPolicyName = dto.PasswordPolicyName
	r.ProtectedEntry = dto.ProtectedEntry
	r.RunCommand = dto.RunCommand
	r.ShiftDoubleClickAction = dto.ShiftDoubleClickAction
	r.Title = dto.Title
	r.Username = dto.Username
	r.URL = dto.URL
	r.UUID = dto.UUID

	return r, nil
}

func saveDB(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}

	var buf bytes.Buffer
	if err := db.Encrypt(&buf); err != nil {
		return fmt.Sprintf("failed to encrypt db: %s", err)
	}

	// Return Uint8Array
	dst := js.Global().Get("Uint8Array").New(buf.Len())
	js.CopyBytesToJS(dst, buf.Bytes())
	return dst
}

func addRecord(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}
	if len(args) != 1 {
		return "invalid arguments: expected (recordJSON)"
	}

	var dto RecordDTO
	if err := json.Unmarshal([]byte(args[0].String()), &dto); err != nil {
		return fmt.Sprintf("json unmarshal error: %s", err)
	}

	record, _ := dto.toRecord()
	db.SetRecord(record)
	return nil
}

func updateRecord(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}
	if len(args) != 2 {
		return "invalid arguments: expected (oldTitle, recordJSON)"
	}

	oldTitle := args[0].String()
	var dto RecordDTO
	if err := json.Unmarshal([]byte(args[1].String()), &dto); err != nil {
		return fmt.Sprintf("json unmarshal error: %s", err)
	}

	record, _ := dto.toRecord()

	if oldTitle != record.Title {
		db.DeleteRecord(oldTitle)
	}
	db.SetRecord(record)
	return nil
}

func deleteRecord(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}
	if len(args) != 1 {
		return "invalid arguments: expected (title)"
	}

	title := args[0].String()
	db.DeleteRecord(title)
	return nil
}
