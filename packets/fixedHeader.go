package packets

import (
	"fmt"
)

type FixedHeader struct {
	messageType byte		
	dup bool				// duplicate delivery flag
	qos byte				// quality of service
	retain bool				// retain flah
	remaingLength uint32	// remaining length including data in the variable header and the payload.
}

func DecodeFixedHeader(buf []byte) *FixedHeader{
	fixedHeader := &FixedHeader{}
	fixedHeader.messageType = buf[0] >> 4
	fixedHeader.dup = buf[0] & 0x08 == 0x08
	fixedHeader.qos = buf[0] & 0x06 >> 1
	fixedHeader.retain = buf[0] & 0x01 == 0x01
	fixedHeader.remaingLength = decodeRemainingLength(buf[1:])
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
	f.messageType, f.dup, f.qos, f.retain, f.remaingLength)
}