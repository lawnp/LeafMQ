package packets

type ErrInvalidFixedHeader struct {}
type ErrInvalidQoS struct {}


func (e *ErrInvalidFixedHeader) Error() string {
	return "Invalid Fixed Header"
}

func (e *ErrInvalidQoS) Error() string {
	return "Invalid QoS"
}

type Subscribtions struct {
	Subscriptions map[string]byte
	OrderedSubscriptions []string
}

func (s *Subscribtions) GetAll() map[string]byte {
	return s.Subscriptions
} 

func (p *Packet) DecodeSubscribe(buf []byte) error {
	if !validFHSubscribe(p.FixedHeader) {
		return &ErrInvalidFixedHeader{}
	}

	buf = p.DecodePacketIdentifier(buf)
	p.Subscriptions = DecodeTopicsSubscribe(buf)
	return nil
}

func (p *Packet) DecodeUnsubscribe(buf []byte) error {
	if !validFHSubscribe(p.FixedHeader) {
		return &ErrInvalidFixedHeader{}
	}
	buf = p.DecodePacketIdentifier(buf)
	p.Subscriptions = DecodeTopicsUnsubscribe(buf)
	return nil
}

// TODO: cleanup this ugly ass code.
func DecodeTopicsSubscribe(buf []byte) *Subscribtions {
	l := len(buf)
	topics := make(map[string]byte)
	topicsOrdered := make([]string, 0)
	for i := 0; i < l; {
		topic, n := DecodeUTF8String(buf[i:])
		qos, err := DecodeQoS(buf[i+int(n)+2])

		//  3 bytes because of 2 bytes for UTF8 string length and 1 byte for qos + length of topic
		i += int(n) + 3

		if err != nil {
			return nil
		}
		topics[topic] = qos
		topicsOrdered = append(topicsOrdered, topic)
	}
	return &Subscribtions{
		topics,
		topicsOrdered,
	}
}

func DecodeTopicsUnsubscribe(buf []byte) *Subscribtions {
	l := len(buf)
	topicsOrdered := make([]string, 0)
	for i := 0; i < l; {
		topic, n := DecodeUTF8String(buf[i:])

		// 3 bytes because of 2 bytes for UTF8 string length + length of topic
		i += int(n) + 2
		topicsOrdered = append(topicsOrdered, topic)
	}
	return &Subscribtions{
		OrderedSubscriptions: topicsOrdered,
	}
}

func DecodeQoS(b byte) (byte, error) {
	qos := b & 0x03
	if (b >> 2 != 0) && (qos > 2) {
		return 0, &ErrInvalidQoS{}
	}
	return qos, nil
}

func validFHSubscribe(fixedHeader *FixedHeader) bool {
	// last four bits of first byte have to be set to 0010 [MQTT-3.8.1-1]
	return fixedHeader.Qos == 1 && 
	!fixedHeader.Dup && !fixedHeader.Retain
}

func (p *Packet) EncodeSuback() []byte {
	// buffer size is: 2 bytes fixed header + 2 bytes varible header + one byte per topic
	buf := make([]byte, 4 + len(p.Subscriptions.OrderedSubscriptions))
	buf[0] = 0x90
	buf[1] = byte(len(buf) - 2)
	buf[2] = byte(p.PacketIdentifier >> 8)
	buf[3] = byte(p.PacketIdentifier & 0xFF)
	for i, topic := range p.Subscriptions.OrderedSubscriptions {
		buf[i+4] = p.Subscriptions.Subscriptions[topic]
	}
	return buf
}

func (p *Packet) EncodeUnsuback() []byte {
	buf := make([]byte, 4)
	buf[0] = 0xB0
	buf[1] = 2
	buf[2] = byte(p.PacketIdentifier >> 8)
	buf[3] = byte(p.PacketIdentifier & 0xFF)
	return buf
}