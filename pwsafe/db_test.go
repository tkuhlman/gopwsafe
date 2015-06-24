package pwsafe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestByteToInt(t *testing.T) {
	var testData = []struct {
		bytes []byte
		value int
	}{
		{bytes: []byte{5}, value: 5},
		{bytes: []byte{5, 5}, value: 1285},
		{bytes: []byte{5, 5, 5}, value: 328965},
		{bytes: []byte{5, 5, 5, 5}, value: 84215045},
		{bytes: []byte{255, 255, 255, 255}, value: 4294967295},
	}

	for _, test := range testData {
		derived := byteToInt(test.bytes)
		assert.Equal(t, test.value, derived)
	}
}
