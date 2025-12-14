package pwsafe

import (
	"testing"
)

func FuzzRecordTitle(f *testing.F) {
	f.Add("Test Title")
	f.Add("")
	f.Add("Another Title with symbols !@#$%^&*()")

	f.Fuzz(func(t *testing.T, title string) {
		r := &Record{Title: title}
		fuzzRoundTrip(t, r, func(r *Record) string { return r.Title }, title)
	})
}

func FuzzRecordUsername(f *testing.F) {
	f.Add("jdoe")
	f.Add("admin")
	f.Add("")

	f.Fuzz(func(t *testing.T, username string) {
		r := &Record{Username: username}
		fuzzRoundTrip(t, r, func(r *Record) string { return r.Username }, username)
	})
}

func FuzzRecordPassword(f *testing.F) {
	f.Add("password123")
	f.Add("CorrectHorseBatteryStaple")
	f.Add("")
	f.Add("Symbols.!@#$%^&*()")

	f.Fuzz(func(t *testing.T, password string) {
		r := &Record{Password: password}
		fuzzRoundTrip(t, r, func(r *Record) string { return r.Password }, password)
	})
}

func FuzzRecordNotes(f *testing.F) {
	f.Add("Some notes here")
	f.Add("Multi\nline\nnotes")
	f.Add("")

	f.Fuzz(func(t *testing.T, notes string) {
		r := &Record{Notes: notes}
		fuzzRoundTrip(t, r, func(r *Record) string { return r.Notes }, notes)
	})
}

func FuzzRecordEmail(f *testing.F) {
	f.Add("test@example.com")
	f.Add("foo+bar@baz.co.uk")
	f.Add("")

	f.Fuzz(func(t *testing.T, email string) {
		r := &Record{Email: email}
		fuzzRoundTrip(t, r, func(r *Record) string { return r.Email }, email)
	})
}

func FuzzRecordURL(f *testing.F) {
	f.Add("https://example.com")
	f.Add("http://localhost:8080")
	f.Add("")

	f.Fuzz(func(t *testing.T, url string) {
		r := &Record{URL: url}
		fuzzRoundTrip(t, r, func(r *Record) string { return r.URL }, url)
	})
}

func fuzzRoundTrip(t *testing.T, r *Record, getter func(*Record) string, original string) {
	// 1. Marshal the record
	data, _, err := r.marshal()
	if err != nil {
		t.Fatalf("Failed to marshal record: %v", err)
	}

	// 2. Unmarshal into a new record
	// We need to simulate the "end of enty" that might be expected if unmarshalRecord doesn't handle single fragments well,
	// but looking at unmarshalRecord, it consumes until recordEndOfEntry or end of slice with error.
	// The marshal() method appends recordEndOfEntry.

	newR := &Record{}
	_, _, err = unmarshalRecord(data, newR)
	if err != nil {
		// If the string is massive, we might hit size limits (though binary.Write handles int sizes).
		// For general strings, it should work.
		// However, if the generation creates data that violates internal constraints (like 2GB strings), verify if that's expected.
		// For now, let's assume no error should occur for reasonable fuzz inputs.
		t.Fatalf("Failed to unmarshal record with input %q: %v", original, err)
	}

	// 3. Verify
	result := getter(newR)
	if result != original {
		t.Errorf("Round trip failed for input %q. Got %q", original, result)
	}
}
