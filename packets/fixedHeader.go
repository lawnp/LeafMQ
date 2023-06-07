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

func DecodeFixedHeader(conn net.Conn) (*FixedHeader, error) {
	buf := make([]byte, 1)
	n, err := conn.Read(buf)

	if err != nil {
		return nil, err
	}
	fmt.Println("Read", n, "bytes from fixed header")
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
	n, err := conn.Read(buf)

	if err != nil {
		return 0, err
	}
	fmt.Println("Read", n, "bytes from byte")
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
