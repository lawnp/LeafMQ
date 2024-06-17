package packets

import (
	"bufio"
	"fmt"
	"io"
)

type FixedHeader struct {
	MessageType     byte
	Dup             bool   // duplicate delivery flag
	Qos             byte   // quality of service
	Retain          bool   // retain flag
	RemainingLength uint32 // remaining length including data in the variable header and the payload.
}

func (f *FixedHeader) Copy() *FixedHeader {
	fixedHeader := *f
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

// Takes in a byteReader and returns decoded fixedHeader. Returns error if we can't read from the connection.
func DecodeFixedHeader(byteReader *bufio.Reader) (*FixedHeader, error) {
	byteBuf, err := byteReader.ReadByte()

	if err != nil {
		return nil, err
	}

	fixedHeader := new(FixedHeader)

	decodeFirstByte(fixedHeader, byteBuf)

	fixedHeader.RemainingLength, err = decodeRemainingLength(byteReader)

	if err != nil {
		return nil, err
	}
	return fixedHeader, nil
}

func decodeFirstByte(fh *FixedHeader, b byte) {
	fh.MessageType = b >> 4
	fh.Dup = b&0x08 == 0x08
	fh.Qos = b&0x06 >> 1
	fh.Retain = b&0x01 == 0x01
}

// decodes the remaining length from the fixed header
func decodeRemainingLength(byteReader io.ByteReader) (uint32, error) {
	var multiplier uint32 = 1
	var value uint32 = 0
	var encodedByte byte
	var err error

	for {
		encodedByte, err = byteReader.ReadByte()
		if err != nil {
			return 0, err
		}
		value += uint32(encodedByte&127) * multiplier
		multiplier *= 128
		if value > 128*128*128 {
			return 0, fmt.Errorf("malformed remaining length")
		}
		if encodedByte&128 == 0 {
			break
		}
	}
	return value, nil
}

