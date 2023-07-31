package packets

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

// MockConn is a mock implementation of the net.Conn interface.
type MockConn struct {
	ReadData []byte
	ReadErr  error
}

func (c *MockConn) Read(buffer []byte) (int, error) {
	if c.ReadErr != nil {
		return 0, c.ReadErr
	}

	if len(c.ReadData) == 0 {
		return 0, errors.New("EOF")
	}

	n := copy(buffer, c.ReadData)
	c.ReadData = c.ReadData[n:]
	return n, nil
}

func (c *MockConn) Write(buffer []byte) (int, error) {
	return 0, nil
}

func (c *MockConn) Close() error {
	return nil
}

func (c *MockConn) LocalAddr() net.Addr {
	return nil
}

func (c *MockConn) RemoteAddr() net.Addr {
	return nil
}

func (c *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

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
			mockConn := &MockConn{ReadData: test.readData}
			value, err := decodeRemainingLength(mockConn)

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
