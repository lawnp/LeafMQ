package packets

const (
	RESERVED       byte = iota // 0 - we use this in packet tests to indicate special-test or all packets.
	CONNECT                    // 1
	CONNACT                    // 2
)

type Packet struct {
	fixedHeader *FixedHeader
	variableHeader *VariableHeader
	payload []byte
}

func (p *Packet) String() string {
	return p.fixedHeader.String()
}

type VariableHeader struct {}