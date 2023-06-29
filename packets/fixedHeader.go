package packets

import (
	"fmt"
	"net"
)

type FixedHeader struct {
	MessageType     byte
	Dup             bool   // duplicate delivery flag
	Qos             byte   // quality of service
	Retain          bool   // retain flag
	RemainingLength uint32 // remaining length including data in the variable header and the payload.
}

func (fh *FixedHeader) Copy() *FixedHeader {
	fixedHeader := *fh
	return &fixedHeader
}

func (f *FixedHeader) Encode() []byte {
	buf := make([]byte, 1)
	buf[0] = f.MessageType << 4
	if f.Dup {
		buf[0] |= 0x08
	}
	buf[0] |= f.Qos << 1
	if f.Retain {
		buf[0] |= 0x01
	}
	buf = append(buf, f.encodeRemainingLength()...)
	return buf
}

func (f *FixedHeader) encodeRemainingLength() []byte {
	var encodedByte byte
	var buffer []byte
	remainingLength := f.RemainingLength

	for {
		encodedByte = byte(remainingLength % 128)
		remainingLength /= 128
		if remainingLength > 0 {
			encodedByte |= 128
		}
		buffer = append(buffer, encodedByte)
		if remainingLength == 0 {
			break
		}
	}
	return buffer
}

func DecodeFixedHeader(conn net.Conn) (*FixedHeader, error) {
	buf := make([]byte, 1)
	_, err := conn.Read(buf)

	if err != nil {
		return nil, err
	}

	fixedHeader := &FixedHeader{}
	fixedHeader.MessageType = buf[0] >> 4
	fixedHeader.Dup = buf[0]&0x08 == 0x08
	fixedHeader.Qos = buf[0] & 0x06 >> 1
	fixedHeader.Retain = buf[0]&0x01 == 0x01
	fixedHeader.RemainingLength = decodeRemainingLength(conn)
	return fixedHeader, nil
}

// decodes the remaining length from the fixed header
func decodeRemainingLength(conn net.Conn) uint32 {
	var multiplier uint32 = 1
	var value uint32 = 0
	var encodedByte byte
	var err error

	for {
		encodedByte, err = decodeByte(conn)
		if err != nil {
			fmt.Println("Error decoding remaining length:", err)
			return 0
		}
		value += uint32(encodedByte&127) * multiplier
		multiplier *= 128
		if multiplier > 128*128*128 {
			fmt.Println("Malformed remaining length")
		}
		if encodedByte&128 == 0 {
			break
		}
	}
	return value
}

func decodeByte(conn net.Conn) (byte, error) {
	buf := make([]byte, 1)
	_, err := conn.Read(buf)

	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (f *FixedHeader) String() string {
	return fmt.Sprintf(
		`Message type: %d
	Dup: %t
	QoS: %d
	Retain: %t
	Remaining length: %d`,
		f.MessageType, f.Dup, f.Qos, f.Retain, f.RemainingLength)
}
