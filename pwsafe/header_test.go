package pwsafe

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/twofish"
)

// buildHeaderField is a helper to construct a single padded header field.
// A field consists of:
// - 4 bytes: length of data (L) (little-endian)
// - 1 byte: field type (T)
// - L bytes: data
// - Padding to make the total entry a multiple of twofish.BlockSize (16 bytes)
func buildHeaderField(fieldType byte, data []byte) []byte {
	dataLength := len(data)

	fieldBuf := new(bytes.Buffer)

	// Write length (4 bytes, little-endian)
	lenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBytes, uint32(dataLength))
	fieldBuf.Write(lenBytes)

	// Write type (1 byte)
	fieldBuf.WriteByte(fieldType)

	// Write data
	fieldBuf.Write(data)

	// Calculate and append padding
	// The total length considered for padding is (length_bytes + type_byte + data_bytes)
	totalLengthSoFar := 4 + 1 + dataLength
	padding := 0
	if totalLengthSoFar%twofish.BlockSize != 0 {
		padding = twofish.BlockSize - (totalLengthSoFar % twofish.BlockSize)
	}

	for i := 0; i < padding; i++ {
		fieldBuf.WriteByte(0x00) // Assuming padding byte is 0x00
	}

	return fieldBuf.Bytes()
}

func TestUnmarshalHeader_MissingEndField(t *testing.T) {
	// Field 1: Version (type 0x00, data {0x0E, 0x03})
	versionData := []byte{0x0E, 0x03}
	versionFieldBytes := buildHeaderField(0x00, versionData)

	// Field 2: Name (type 0x09, data "test")
	nameData := []byte("test")
	nameFieldBytes := buildHeaderField(0x09, nameData)

	// Concatenate fields - NO END field
	headerBytes := append(versionFieldBytes, nameFieldBytes...)

	// var h header // h is not used when only checking for error
	_, _, _, err := UnmarshalHeader(headerBytes) // Use _, _, _, err as return values match this

	assert.NotNil(t, err, "UnmarshalHeader should return an error for missing END field")
	// Based on header.go, the error for running out of data before END is found:
	assert.Equal(t, "no END field found when UnMarshaling", err.Error(), "Error message mismatch")
}

func TestUnmarshalHeader_UnknownFieldType(t *testing.T) {
	// Field 1: Version (type 0x00, data {0x0E, 0x03})
	versionData := []byte{0x0E, 0x03}
	versionFieldBytes := buildHeaderField(0x00, versionData)

	// Field 2: Unknown Field (type 0xFE, data {0x01})
	unknownFieldData := []byte{0x01}
	unknownFieldBytes := buildHeaderField(0xFE, unknownFieldData)

	// Field 3: END (type 0xFF, data nil)
	endFieldBytes := buildHeaderField(0xFF, nil)

	// Concatenate fields
	headerBytes := append(versionFieldBytes, unknownFieldBytes...)
	headerBytes = append(headerBytes, endFieldBytes...)

	// var h header // h is not used when only checking for error
	_, _, _, err := UnmarshalHeader(headerBytes)

	assert.NotNil(t, err, "UnmarshalHeader should return an error for unknown field type")
	// Based on header.go, the error for unknown field type:
	expectedError := fmt.Sprintf("encountered unknown Header Field type - %v", 0xFE)
	assert.Equal(t, expectedError, err.Error(), "Error message mismatch")
}

func TestUnmarshalHeader_FieldLengthExceedsData(t *testing.T) {
	// Field 1: Version (type 0x00, data {0x0E, 0x03})
	versionFieldBytes := buildHeaderField(0x00, []byte{0x0E, 0x03})

	// Field 2: Problematic field (e.g., Description type 0x0a)
	// Declare length as 255, but provide only 3 bytes of data.
	fieldTypeProblem := byte(0x0a) // DescriptionField
	declaredLength := uint32(255)
	actualData := []byte{0x01, 0x02, 0x03}

	malformedFieldHeader := new(bytes.Buffer)
	lenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBytes, declaredLength)
	malformedFieldHeader.Write(lenBytes)
	malformedFieldHeader.WriteByte(fieldTypeProblem)
	// DO NOT append padding here as UnmarshalHeader will try to read data first based on declaredLength

	headerBytes := append(versionFieldBytes, malformedFieldHeader.Bytes()...)
	headerBytes = append(headerBytes, actualData...) // Append only the short actual data

	// UnmarshalHeader will try to read `fieldStart+5 : fieldStart+fieldLength+5`.
	// If fieldLength is 255, this will be fieldStart+5 : fieldStart+255+5.
	// This is expected to panic due to out-of-bounds slice access.
	// UnmarshalHeader should now return an error instead of panicking
	_, _, _, err := UnmarshalHeader(headerBytes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid field length", "Error should indicate invalid field length")
}

func TestUnmarshalHeader_EmptyOrTooShortInput(t *testing.T) {
	t.Run("Empty byte slice", func(t *testing.T) {
		emptyData := []byte{}
		_, _, _, err := UnmarshalHeader(emptyData)
		assert.NotNil(t, err, "Should return error with empty data slice")
		assert.Equal(t, "no END field found when UnMarshaling", err.Error())
	})

	t.Run("Too short for field header", func(t *testing.T) {
		shortData := []byte{0x01, 0x02, 0x03} // Only 3 bytes
		_, _, _, err := UnmarshalHeader(shortData)
		assert.NotNil(t, err, "Should return error with data slice too short for a field header")
		assert.Equal(t, "no END field found when UnMarshaling", err.Error())
	})
}
