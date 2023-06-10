package packets

import (
	"fmt"
	"net"
)

const (
	RESERVED       byte = iota // 0
	CONNECT                    // 1
	CONNACK                    // 2
	PUBLISH                    // 3
	PUBACK                     // 4
	PUBREC                     // 5
	PUBREL                     // 6
	PUBCOMP                    // 7
	SUBSCRIBE                  // 8
	SUBACK                     // 9
	UNSUBSCRIBE                // 10
	UNSUBACK                   // 11
	PINGREQ                    // 12
	PINGRES                   // 13
	DISCONNECT                 // 14
	AUTH                       // 15
)

type Packet struct {
	FixedHeader *FixedHeader
	ConnectOptions *ConnectOptions
	Subscriptions *Subscribtions
	PublishTopic string
	PacketIdentifier uint16
	Payload []byte
}

func (p *Packet) String() string {
	return p.FixedHeader.String()
}

func ParsePacket(fh *FixedHeader, conn net.Conn) (*Packet, error) {
	packet := new(Packet)
	packet.FixedHeader = fh
	var err error	

	buf := make([]byte, fh.RemainingLength)
	_, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}

	switch packet.FixedHeader.MessageType {
	case CONNECT:
		packet.ConnectOptions, err = DecodeConnect(buf)
	case SUBSCRIBE:
		err = packet.DecodeSubscribe(buf)
	case UNSUBSCRIBE:
		err = packet.DecodeUnsubscribe(buf)
	case PUBLISH:
		err = packet.DecodePublish(buf)
	case PINGREQ:
	case DISCONNECT:
	default:
		fmt.Println("Unknown packet type: ", packet.FixedHeader.MessageType)
	}

	if err != nil {
		panic(err)
	}
	return packet, err
}

func EncodePingresp() []byte {
	buffer := make([]byte, 2)
	buffer[0] = 13 << 4
	buffer[1] = 0
	return buffer
}
