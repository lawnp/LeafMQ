package packets

import (
	"fmt"
	"net"
)

const (
	RESERVED       byte = iota // 0
	CONNECT                    // 1
	CONNACK                   // 2
)

type Packet struct {
	fixedHeader *FixedHeader
	ConnectOptions *ConnectOptions
	// payload []byte
}

func (p *Packet) String() string {
	return p.fixedHeader.String()
}

func ParsePacket(fh *FixedHeader, conn net.Conn) (*Packet, error) {
	packet := new(Packet)
	packet.fixedHeader = fh
	var err error	

	buf := make([]byte, fh.RemainingLength)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	fmt.Println("Read", n, "bytes")
	switch packet.fixedHeader.MessageType {
	case CONNECT:
		packet.ConnectOptions, err = DecodeConnect(buf)
	}

	if err != nil {
		panic(err)
	}
	return packet, err
}
