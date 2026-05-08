package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
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

	for uuidHex, rec := range db.Records {
		items = append(items, Item{
			UUID:  uuidHex,
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

func getRecord(this js.Value, args []js.Value) any {
	if db == nil {
		return "database not open"
	}
	if len(args) != 1 {
		return "invalid arguments: expected (uuid)"
	}

	uuidHex := args[0].String()
	rec, ok := db.Records[uuidHex]
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

// UpdateRecordFields creates or updates a record.
// Args: uuid, field1, value1, ...
// Pass empty string for uuid to create a new record (UUID is generated and returned).
// Supported fields: Title, Group, Username, Password, URL, Notes
// Title and Password must be non-empty in the final record.
func updateRecordFields(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}
	if len(args) < 3 || len(args)%2 == 0 {
		return "invalid arguments: expected (uuid, field, value, ...)"
	}
	uuidHex := args[0].String()
	var rec pwsafe.Record
	if uuidHex != "" {
		var ok bool
		rec, ok = db.Records[uuidHex]
		if !ok {
			return "record not found"
		}
	}
	for i := 1; i+1 < len(args); i += 2 {
		field, value := args[i].String(), args[i+1].String()
		switch field {
		case "Title":
			rec.Title = value
		case "Group":
			rec.Group = value
		case "Username":
			rec.Username = value
		case "Password":
			if value != rec.Password && rec.Password != "" {
				rec.PasswordHistory = pushPasswordHistory(rec.PasswordHistory, rec.Password)
			}
			rec.Password = value
		case "URL":
			rec.URL = value
		case "Notes":
			rec.Notes = value
		default:
			return fmt.Sprintf("unknown field: %s", field)
		}
	}
	if rec.Title == "" {
		return "Title is required"
	}
	if rec.Password == "" {
		return "Password is required"
	}
	return db.SetRecord(rec)
}

// UpdateDBFields applies a field/value patch to the database header.
// Args: field1, value1, field2, value2, ...
// Supported fields: Name, Description, LastSaveUser
func updateDBFields(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}
	if len(args) < 2 || len(args)%2 != 0 {
		return "invalid arguments: expected (field, value, ...)"
	}
	for i := 0; i+1 < len(args); i += 2 {
		field, value := args[i].String(), args[i+1].String()
		switch field {
		case "Name":
			db.Header.Name = value
		case "Description":
			db.Header.Description = value
		case "LastSaveUser":
			db.Header.LastSaveUser = []byte(value)
		default:
			return fmt.Sprintf("unknown field: %s", field)
		}
	}
	return nil
}

// pushPasswordHistory appends oldPassword to the pwsafe password history string.
// Format: fmmnn[T(8hex)L(4hex)P]...
//   f='1' enabled, mm=max entries (hex), nn=count (hex)
//   each entry: 8-char hex Unix timestamp, 4-char hex pw length, password
func pushPasswordHistory(current, oldPassword string) string {
	type entry struct {
		ts int64
		pw string
	}
	enabled := true
	maxEntries := 10
	var entries []entry

	if len(current) >= 5 {
		enabled = current[0] == '1'
		if m, err := strconv.ParseInt(current[1:3], 16, 64); err == nil && m > 0 {
			maxEntries = int(m)
		}
		count := 0
		if c, err := strconv.ParseInt(current[3:5], 16, 64); err == nil {
			count = int(c)
		}
		pos := 5
		for i := 0; i < count; i++ {
			if pos+12 > len(current) {
				break
			}
			ts, err := strconv.ParseInt(current[pos:pos+8], 16, 64)
			if err != nil {
				break
			}
			pos += 8
			l, err := strconv.ParseInt(current[pos:pos+4], 16, 64)
			if err != nil {
				break
			}
			pos += 4
			if pos+int(l) > len(current) {
				break
			}
			entries = append(entries, entry{ts, current[pos : pos+int(l)]})
			pos += int(l)
		}
	}

	if !enabled {
		return current
	}

	entries = append(entries, entry{time.Now().Unix(), oldPassword})
	for len(entries) > maxEntries {
		entries = entries[1:]
	}

	var sb strings.Builder
	if enabled {
		sb.WriteByte('1')
	} else {
		sb.WriteByte('0')
	}
	sb.WriteString(fmt.Sprintf("%02x%02x", maxEntries, len(entries)))
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("%08x%04x%s", e.ts, len(e.pw), e.pw))
	}
	return sb.String()
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
	uuids := db.Search(query, namesOnly)
	jsonData, err := json.Marshal(uuids)
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
	js.Global().Set("UpdateRecordFields", js.FuncOf(updateRecordFields))
	js.Global().Set("deleteRecord", js.FuncOf(deleteRecord))
	js.Global().Set("UpdateDBFields", js.FuncOf(updateDBFields))
	js.Global().Set("searchRecords", js.FuncOf(searchRecords))

	fmt.Println("WASM initialized")
	<-c
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

func deleteRecord(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return "database not open"
	}
	if len(args) != 1 {
		return "invalid arguments: expected (uuid)"
	}

	db.DeleteRecord(args[0].String())
	return nil
}
