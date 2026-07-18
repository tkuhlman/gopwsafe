package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"syscall/js"
	"time"

	"github.com/tkuhlman/gopwsafe/pwsafe"
)

var db *pwsafe.V3

func openDB(this js.Value, args []js.Value) any {
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

func getDBData(this js.Value, args []js.Value) any {
	if db == nil {
		return "database not open"
	}

	type Item struct {
		UUID  string `json:"uuid"`
		Title string `json:"title"`
		Group string `json:"group"`
	}

	var items []Item

	for _, rec := range db.Records {
		uuidStr := fmt.Sprintf("%x", rec.UUID)
		items = append(items, Item{
			UUID:  uuidStr,
			Title: rec.Title,
			Group: rec.Group,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Group != items[j].Group {
			return items[i].Group < items[j].Group
		}
		if items[i].Title != items[j].Title {
			return items[i].Title < items[j].Title
		}
		return items[i].UUID < items[j].UUID
	})

	jsonData, err := json.Marshal(items)
	if err != nil {
		return fmt.Sprintf("json marshal error: %s", err)
	}

	return string(jsonData)
}

func getRecord(this js.Value, args []js.Value) any {
	if db == nil {
		return "database not open"
	}
	if len(args) != 1 {
		return "invalid arguments: expected (uuid)"
	}

	uuidStr := args[0].String()
	bytes, err := hex.DecodeString(uuidStr)
	if err != nil || len(bytes) != 16 {
		return "invalid uuid format"
	}
	var uuidBytes [16]byte
	copy(uuidBytes[:], bytes)

	rec, ok := db.Records[uuidBytes]
	if !ok {
		return "record not found"
	}

	jsonData, err := json.Marshal(rec)
	if err != nil {
		return fmt.Sprintf("json marshal error: %s", err)
	}

	return string(jsonData)
}

func createDatabase(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return "invalid arguments: expected (password)"
	}
	password := args[0].String()

	newDB := pwsafe.NewV3("", password)
	db = newDB
	return nil
}

func getDBInfo(this js.Value, args []js.Value) any {
	if db == nil {
		return "database not open"
	}
	// Return header info
	// db.Header contains the info.
	type DBInfo struct {
		Version     string `json:"version"`
		UUID        string `json:"uuid"`
		Name        string `json:"name"`
		Description string `json:"description"`
		What        string `json:"what"`
		When        string `json:"when"`
		Who         string `json:"who"`
	}

	// UUID to string
	uuidStr := fmt.Sprintf("%x", db.Header.UUID)

	// Map of known versions
	versionMap := map[uint16]string{
		0x0300: "3.01",
		0x0301: "3.03",
		0x0302: "3.09",
		0x0303: "3.12",
		0x0304: "3.13",
		0x0305: "3.14",
		0x0306: "3.19",
		0x0307: "3.22",
		0x0308: "3.25",
		0x0309: "3.26",
		0x030A: "3.28",
		0x030B: "3.29",
		0x030C: "3.29",
		0x030D: "3.30",
		0x030E: "3.47",
		0x030F: "3.68",
		0x0310: "3.69",
	}

	versionVal := binary.LittleEndian.Uint16(db.Header.Version[:])
	versionStr := versionMap[versionVal]
	if versionStr == "" {
		versionStr = fmt.Sprintf("Format 0x%04x", versionVal)
	} else {
		versionStr = "v" + versionStr
	}

	info := DBInfo{
		Version:     versionStr,
		UUID:        uuidStr,
		Name:        db.Header.Name,
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

func updateDBInfo(this js.Value, args []js.Value) any {
	if db == nil {
		return "database not open"
	}
	if len(args) != 2 {
		return "invalid arguments: expected (name, description)"
	}

	name := args[0].String()
	description := args[1].String()

	db.Header.Name = name
	db.Header.Description = description

	return nil
}

func searchRecords(this js.Value, args []js.Value) any {
	if db == nil {
		return "database not open"
	}
	if len(args) != 2 {
		return "invalid arguments: expected (query, namesOnly)"
	}
	query := args[0].String()
	namesOnly := args[1].Bool()
	titles := db.Search(query, namesOnly)
	jsonData, err := json.Marshal(titles)
	if err != nil {
		return fmt.Sprintf("json marshal error: %s", err)
	}
	return string(jsonData)
}

func getSuggestion(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return ""
	}
	if len(args) != 2 {
		return ""
	}
	field := args[0].String()
	prefix := args[1].String()
	if prefix == "" {
		return ""
	}

	prefixLower := strings.ToLower(prefix)
	freq := make(map[string]int)
	for _, rec := range db.Records {
		var val string
		switch field {
		case "group":
			val = rec.Group
		case "username":
			val = rec.Username
		default:
			return ""
		}
		if val != "" {
			freq[val]++
		}
	}

	var matches []string
	for val := range freq {
		if strings.HasPrefix(strings.ToLower(val), prefixLower) {
			matches = append(matches, val)
		}
	}
	if len(matches) == 0 {
		return ""
	}

	sort.Slice(matches, func(i, j int) bool {
		fi, fj := freq[matches[i]], freq[matches[j]]
		if fi != fj {
			return fi > fj
		}
		return strings.ToLower(matches[i]) < strings.ToLower(matches[j])
	})
	return matches[0]
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
	js.Global().Set("updateDBInfo", js.FuncOf(updateDBInfo))
	js.Global().Set("searchRecords", js.FuncOf(searchRecords))
	js.Global().Set("getSuggestion", js.FuncOf(getSuggestion))

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

func saveDB(this js.Value, args []js.Value) any {
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

func addRecord(this js.Value, args []js.Value) any {
	if db == nil {
		return `{"error":"database not open"}`
	}
	if len(args) != 1 {
		return `{"error":"invalid arguments: expected (recordJSON)"}`
	}

	var dto RecordDTO
	if err := json.Unmarshal([]byte(args[0].String()), &dto); err != nil {
		return fmt.Sprintf(`{"error":"json unmarshal error: %s"}`, err)
	}

	record, _ := dto.toRecord()
	uuidBytes := db.SetRecord(record)
	return fmt.Sprintf(`{"uuid":"%x"}`, uuidBytes)
}

func updateRecord(this js.Value, args []js.Value) any {
	if db == nil {
		return `{"error":"database not open"}`
	}
	if len(args) != 2 {
		return `{"error":"invalid arguments: expected (oldUUID, recordJSON)"}`
	}

	oldUUIDStr := args[0].String()
	bytes, err := hex.DecodeString(oldUUIDStr)
	if err != nil || len(bytes) != 16 {
		return `{"error":"invalid oldUUID format"}`
	}
	var oldUUID [16]byte
	copy(oldUUID[:], bytes)

	var dto RecordDTO
	if err := json.Unmarshal([]byte(args[1].String()), &dto); err != nil {
		return fmt.Sprintf(`{"error":"json unmarshal error: %s"}`, err)
	}

	record, _ := dto.toRecord()

	if oldUUID != record.UUID {
		db.DeleteRecord(oldUUID)
	}
	db.SetRecord(record)
	return `{"success":true}`
}

func deleteRecord(this js.Value, args []js.Value) any {
	if db == nil {
		return `{"error":"database not open"}`
	}
	if len(args) != 1 {
		return `{"error":"invalid arguments: expected (uuid)"}`
	}

	uuidStr := args[0].String()
	bytes, err := hex.DecodeString(uuidStr)
	if err != nil || len(bytes) != 16 {
		return `{"error":"invalid uuid format"}`
	}
	var uuidBytes [16]byte
	copy(uuidBytes[:], bytes)

	db.DeleteRecord(uuidBytes)
	return `{"success":true}`
}
