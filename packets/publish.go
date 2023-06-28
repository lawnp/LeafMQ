package packets

func (p *Packet) DecodePublish(buf []byte) error {
	var n uint16
	p.PublishTopic, n = DecodeUTF8String(buf)
	if p.FixedHeader.Qos > 0 {
		p.DecodePacketIdentifier(buf[n + 2:])
		n += 2
	}
	p.Payload = buf[n+2:]
	return nil
}

func (p *Packet) DecodePuback(buf []byte) error {
	p.DecodePacketIdentifier(buf)
	return nil
}

func (p *Packet) EncodePublish() []byte {
	var buffer []byte
	buffer = append(buffer, p.FixedHeader.Encode()...)
	buffer = append(buffer, EncodeUTF8String(p.PublishTopic)...)

	if p.FixedHeader.Qos > 0 {
		buffer = append(buffer, EncodePacketIdentifier(p.PacketIdentifier)...)
	}

	buffer = append(buffer, p.Payload...)
	return buffer
}

func (p *Packet) EncodePuback() []byte {
	var buffer []byte
	packet := new(Packet)
	packet.FixedHeader = new(FixedHeader)
	packet.FixedHeader.MessageType = PUBACK
	packet.FixedHeader.RemainingLength = 2
	buffer = append(buffer, packet.FixedHeader.Encode()...)
	buffer = append(buffer, EncodePacketIdentifier(p.PacketIdentifier)...)
	return buffer
}

func (p Packet) EncodeResp() []byte {
	var buffer []byte
	buffer = append(buffer, p.FixedHeader.Encode()...)
	buffer = append(buffer, EncodePacketIdentifier(p.PacketIdentifier)...)
	return buffer
}

func BuildResp(packet *Packet, messageType byte) *Packet {
	resp := new(Packet)
	resp.FixedHeader = new(FixedHeader)
	resp.FixedHeader.MessageType = messageType
	resp.FixedHeader.RemainingLength = 2
	resp.PacketIdentifier = packet.PacketIdentifier
	return resp
}
