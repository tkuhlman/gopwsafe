package pwsafe

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecord_PasswordExpiryInterval(t *testing.T) {
	t.Run("Valid Interval", func(t *testing.T) {
		r := &Record{}
		// 90 days
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, 90)

		err := r.setField(recordPasswordExpiryInterval, data)
		assert.NoError(t, err)
		assert.Equal(t, uint32(90), r.PasswordExpiryInterval)

		// Marshal
		marshaled, _, err := r.marshal()
		assert.NoError(t, err)
		assert.Contains(t, string(marshaled), string(data), "Marshaled data should contain interval")
	})

	t.Run("Invalid Interval - Too Logical Large", func(t *testing.T) {
		r := &Record{}
		// 4000 days (Max is 3650)
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, 4000)

		err := r.setField(recordPasswordExpiryInterval, data)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), r.PasswordExpiryInterval, "Should default to 0 if > 3650")
	})

	t.Run("Marshal Invalid Interval", func(t *testing.T) {
		r := &Record{}
		r.Title = "Test"
		r.Password = "Test"
		r.PasswordExpiryInterval = 5000 // Manually set invalid value

		_, _, err := r.marshal()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum")
	})
}

func TestRecord_OwnSymbolsForPassword(t *testing.T) {
	t.Run("Set and Get OwnSymbolsForPassword", func(t *testing.T) {
		r := &Record{}
		symbols := "!@#$%^&*()"
		
		err := r.setField(recordOwnSymbolsForPassword, []byte(symbols))
		assert.NoError(t, err)
		assert.Equal(t, symbols, r.OwnSymbolsForPassword)
	})

	t.Run("Marshal and Unmarshal OwnSymbolsForPassword", func(t *testing.T) {
		r1 := &Record{
			Title:                 "Test Entry",
			Username:              "testuser",
			Password:              "testpass",
			OwnSymbolsForPassword: "!@#$%^&*()-_=+[]{}",
		}

		// Marshal the record
		marshaled, _, err := r1.marshal()
		assert.NoError(t, err)
		assert.NotNil(t, marshaled)

		// Verify the field is in the marshaled data
		// The marshaled data should contain our symbols
		assert.Contains(t, string(marshaled), r1.OwnSymbolsForPassword)
	})

	t.Run("Empty OwnSymbolsForPassword", func(t *testing.T) {
		r := &Record{
			Title:                 "Test Entry",
			Username:              "testuser",
			Password:              "testpass",
			OwnSymbolsForPassword: "",
		}

		// Marshal should work with empty string
		marshaled, _, err := r.marshal()
		assert.NoError(t, err)
		assert.NotNil(t, marshaled)
	})

	t.Run("Record Equality with OwnSymbolsForPassword", func(t *testing.T) {
		r1 := &Record{
			Title:                 "Test",
			OwnSymbolsForPassword: "!@#$",
		}
		r2 := &Record{
			Title:                 "Test",
			OwnSymbolsForPassword: "!@#$",
		}
		r3 := &Record{
			Title:                 "Test",
			OwnSymbolsForPassword: "different",
		}

		// r1 and r2 should be equal
		equal, err := r1.Equal(*r2, true)
		assert.NoError(t, err)
		assert.True(t, equal)

		// r1 and r3 should not be equal
		equal, err = r1.Equal(*r3, true)
		assert.False(t, equal)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OwnSymbolsForPassword")
	})

	t.Run("UTF-8 Symbols", func(t *testing.T) {
		r := &Record{}
		symbols := "§±¿×÷"
		
		err := r.setField(recordOwnSymbolsForPassword, []byte(symbols))
		assert.NoError(t, err)
		assert.Equal(t, symbols, r.OwnSymbolsForPassword)
	})
}
