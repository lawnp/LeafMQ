package packets

import (
	"fmt"
)

type FixedHeader struct {
	MessageType byte		
	Dup bool				// duplicate delivery flag
	Qos byte				// quality of service
	Retain bool				// retain flag
	RemaingLength uint32	// remaining length including data in the variable header and the payload.
}

func DecodeFixedHeader(buf []byte) *FixedHeader {
	fixedHeader := &FixedHeader{}
	fixedHeader.MessageType = buf[0] >> 4
	fixedHeader.Dup = buf[0] & 0x08 == 0x08
	fixedHeader.Qos = buf[0] & 0x06 >> 1
	fixedHeader.Retain = buf[0] & 0x01 == 0x01
	fixedHeader.RemaingLength = decodeRemainingLength(buf[1:])
	return fixedHeader
}

// decodes the remaining length from the fixed header
func decodeRemainingLength(buf []byte) uint32 {
	var value uint32
	var multiplier uint32 = 1
	for i := 0; i < len(buf); i++ {
		value += uint32(buf[i] & 127) * multiplier
		multiplier *= 128
		if buf[i] & 128 == 0 {
			break
		}
	}
	return value
}

func (f *FixedHeader) String() string {
	return fmt.Sprintf(
	`Message type: %d
	Dup: %t
	QoS: %d
	Retain: %t
	Remaining length: %d`,
	f.MessageType, f.Dup, f.Qos, f.Retain, f.RemaingLength)
}