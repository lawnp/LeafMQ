package packets

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)


func TestDecodeRemainingLength(t *testing.T) {
	tests := []struct {
		name          string
		readData      []byte
		expectedValue uint32
		expectedError error
	}{
		{
			name:          "Valid remaining length - single byte",
			readData:      []byte{0x05},
			expectedValue: 5,
			expectedError: nil,
		},
		{
			name:          "Valid remaining length - multiple bytes part 1",
			readData:      []byte{0x80, 0x01},
			expectedValue: 128,
			expectedError: nil,
		},
		{
			name:          "Valid remaining length - multiple bytes part 2",
			readData:      []byte{0x80, 0x80, 0x01},
			expectedValue: 16384,
			expectedError: nil,
		},
		{
			name:          "Valid remaining length - multiple bytes part 2",
			readData:      []byte{0x80, 0x80, 0x80, 0x01},
			expectedValue: 2097152,
			expectedError: nil,
		},
		{
			name:          "Invalid remaining length - exceeds maximum",
			readData:      []byte{0x80, 0x80, 0x80, 0x80, 0x01},
			expectedValue: 0,
			expectedError: fmt.Errorf("malformed remaining length"),
		},
		{
			name:          "Unexpected EOF",
			readData:      []byte{},
			expectedValue: 0,
			expectedError: errors.New("EOF"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			byteBuffer:= bytes.NewBuffer(test.readData)
			value, err := decodeRemainingLength(byteBuffer)

			if err != nil {
				if test.expectedError == nil {
					t.Errorf("Unexpected error: %v", err)
				} else if err.Error() != test.expectedError.Error() {
					t.Errorf("Expected error: %v, but got: %v", test.expectedError, err)
				}
			} else {
				if test.expectedError != nil {
					t.Errorf("Expected error: %v, but got nil", test.expectedError)
				}
				if value != test.expectedValue {
					t.Errorf("Expected value: %d, but got: %d", test.expectedValue, value)
				}
			}
		})
	}
}

func TestFixedHeader_Encode(t *testing.T) {
	tests := []struct {
		name         string
		fixedHeader  *FixedHeader
		expectedData []byte
	}{
		{
			name: "FixedHeader with all fields set",
			fixedHeader: &FixedHeader{
				MessageType:     1,
				Dup:             true,
				Qos:             2,
				Retain:          true,
				RemainingLength: 321,
			},
			expectedData: []byte{0x1D, 0xC1, 0x02},
		},
		{
			name: "FixedHeader some fields set",
			fixedHeader: &FixedHeader{
				MessageType:     4,
				Dup:             true,
				Qos:             1,
				Retain:          false,
				RemainingLength: 512,
			},
			expectedData: []byte{0x4A, 0x80, 0x04},
		},
		{
			name: "FixedHeader with minimum values",
			fixedHeader: &FixedHeader{
				MessageType:     0x0, // RESERVED packet type, go enough for testing
				Dup:             false,
				Qos:             0,
				Retain:          false,
				RemainingLength: 0,
			},
			expectedData: []byte{0x00, 0x00},
		},
		// Add more test cases for different scenarios.
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := test.fixedHeader.Encode()

			if !bytes.Equal(data, test.expectedData) {
				t.Errorf("Expected data: %v, but got: %v", test.expectedData, data)
			}
		})
	}
}
