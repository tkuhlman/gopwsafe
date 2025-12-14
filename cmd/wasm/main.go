package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"syscall/js"

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

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("openDB", js.FuncOf(openDB))
	js.Global().Set("getDBData", js.FuncOf(getDBData))
	js.Global().Set("getRecord", js.FuncOf(getRecord))
	// js.Global().Set("saveDB", js.FuncOf(saveDB))
	// js.Global().Set("addRecord", js.FuncOf(addRecord))
	// js.Global().Set("updateRecord", js.FuncOf(updateRecord))

	fmt.Println("WASM initialized")
	<-c
}
