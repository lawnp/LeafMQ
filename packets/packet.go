package packets

const (
	RESERVED       byte = iota // 0
	CONNECT                    // 1
	CONNACK                   // 2
)

type Packet struct {
	fixedHeader *FixedHeader
	ConnectOptions *ConnectOptions
	payload []byte
}

func (p *Packet) String() string {
	return p.fixedHeader.String()
}

func ParsePacket(buffer []byte) *Packet {
	packet := new(Packet)
	var err error	

	switch packet.fixedHeader.MessageType {
	case CONNECT:
		
	}

	if err != nil {
		panic(err)
	}
	return packet
}

func (p *Packet) ParseFixedHeader(buffer []byte) {
	p.fixedHeader = DecodeFixedHeader(buffer)
}


