package packets

func (p *Packet) DecodePublish(buf []byte) error {
	var n uint16
	p.PublishTopic, n = DecodeUTF8String(buf)
	p.DecodePacketIdentifier(buf[n:])
	p.Payload = buf[n+2:]
	return nil
}

// reassembling the packet is a bit useless, since we just decoded it.
// but I don't want to save the whole packet in memory again.
// maybe I should... who knows.
func (p *Packet) EncodePublish() []byte {
	var buffer []byte
	buffer = append(buffer, p.FixedHeader.Encode()...)
	buffer = append(buffer, EncodeUTF8String(p.PublishTopic)...)
	//buffer = append(buffer, EncodePacketIdentifier(p.PacketIdentifier)...)
	buffer = append(buffer, p.Payload...)
	return buffer
}