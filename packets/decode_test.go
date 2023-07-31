package packets

import (
	"testing"
	"bytes"
	"reflect"
	"strings"
)

func TestDecodeUTF8String(t *testing.T) {
	tests := []struct {
		name            string
		buffer          []byte
		expectedString  string
		expectedLength  uint16
		expectedHasData bool
	}{
		{
			name:            "Empty buffer",
			buffer:          []byte{},
			expectedString:  "",
			expectedLength:  0,
			expectedHasData: false,
		},
		{
			name:            "Valid UTF-8 string",
			buffer:          []byte{0x00, 0x05, 0x48, 0x65, 0x6C, 0x6C, 0x6F},
			expectedString:  "Hello",
			expectedLength:  5,
			expectedHasData: true,
		},
		{
			name:            "Invalid buffer - length greater than actual data",
			buffer:          []byte{0x00, 0x06, 0x48, 0x65, 0x6C, 0x6C},
			expectedString:  "",
			expectedLength:  0,
			expectedHasData: false,
		},
		{
			name:            "Invalid buffer - no data",
			buffer:          []byte{0x00, 0x00},
			expectedString:  "",
			expectedLength:  0,
			expectedHasData: false,
		},
		// Add more test cases for different scenarios.
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			str, length := DecodeUTF8String(test.buffer)

			if str != test.expectedString {
				t.Errorf("Expected string: %q, but got: %q", test.expectedString, str)
			}

			if length != test.expectedLength {
				t.Errorf("Expected length: %d, but got: %d", test.expectedLength, length)
			}

			if (len(str) > 0) != test.expectedHasData {
				t.Errorf("Expected has data: %v, but got: %v", test.expectedHasData, len(str) > 0)
			}
		})
	}
}

func TestPacket_DecodePacketIdentifier(t *testing.T) {
	tests := []struct {
		name                 string
		buffer               []byte
		expectedPacketID     uint16
		expectedRemainingBuf []byte
	}{
		{
			name:                 "Valid packet identifier",
			buffer:               []byte{0x12, 0x34, 0x56, 0x78},
			expectedPacketID:     0x1234,
			expectedRemainingBuf: []byte{0x56, 0x78},
		},
		{
			name:                 "Packet identifier with minimum value",
			buffer:               []byte{0x00, 0x01, 0x02, 0x03},
			expectedPacketID:     0x0001,
			expectedRemainingBuf: []byte{0x02, 0x03},
		},
		{
			name:                 "Packet identifier with maximum value",
			buffer:               []byte{0xFF, 0xFF, 0xAA, 0xBB},
			expectedPacketID:     0xFFFF,
			expectedRemainingBuf: []byte{0xAA, 0xBB},
		},
		{
			name:                 "Insufficient data in buffer",
			buffer:               []byte{0x11},
			expectedPacketID:     0x00,
			expectedRemainingBuf: []byte{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			packet := &Packet{}
			remainingBuf := packet.DecodePacketIdentifier(test.buffer)

			if packet.PacketIdentifier != test.expectedPacketID {
				t.Errorf("Expected packet identifier: 0x%X, but got: 0x%X", test.expectedPacketID, packet.PacketIdentifier)
			}

			if !bytes.Equal(remainingBuf, test.expectedRemainingBuf) {
				t.Errorf("Expected remaining buffer: %v, but got: %v", test.expectedRemainingBuf, remainingBuf)
			}
		})
	}
}

func TestEncodeUTF8String(t *testing.T) {
	firstTwoBytes := []byte{0xFF, 0xFF}
	longBytes := bytes.Repeat([]byte{0x61}, 65535)
	finalBytes := append(firstTwoBytes, longBytes...)
	tests := []struct {
		name           string
		inputString    string
		expectedResult []byte
	}{
		{
			name:           "Empty string",
			inputString:    "",
			expectedResult: []byte{0x00, 0x00},
		},
		{
			name:           "Short string",
			inputString:    "Hello",
			expectedResult: []byte{0x00, 0x05, 0x48, 0x65, 0x6C, 0x6C, 0x6F},
		},
		{
			name:           "String with non-ASCII characters",
			inputString:    "こんにちは",
			expectedResult: []byte{0x00, 0x0F, 0xE3, 0x81, 0x93, 0xE3, 0x82, 0x93, 0xE3, 0x81, 0xAB, 0xE3, 0x81, 0xA1, 0xE3, 0x81, 0xAF},
		},
		{
			name:           "String with maximum length",
			inputString:    strings.Repeat("a", 65535),
			// MAX LENGTH = 2^16 - 1 = 65535
			expectedResult: finalBytes,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := EncodeUTF8String(test.inputString)

			if !reflect.DeepEqual(result, test.expectedResult) {
				t.Errorf("Expected: %v, but got: %v", test.expectedResult, result)
			}
		})
	}
}