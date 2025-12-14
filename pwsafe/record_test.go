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
