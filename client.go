package nixmq

import (
	"fmt"
	"net"

	"github.com/LanPavletic/nixMQ/packets"
)

type Client struct {
	Propreties *Propreties
	Conn       net.Conn
	Session    *Session
	isClosed   bool
	Broker     *Broker
}

type Propreties struct {
	ProtocolLevel byte
	ClientID      string
	Username      string
	Password      string
	WillTopic     string
	WillMessage   string
	WillRetain    bool
	WillQoS       byte
	CleanSession  bool
	Keepalive     uint16
}

type Session struct {
	PendingPackets map[uint16]*packets.Packet
	Subscriptions *Subscriptions
}

func NewSession() *Session {
	return &Session{
		make(map[uint16]*packets.Packet),
		newSubscriptions(),
	}
}

func (s *Session) Get(packetID uint16) (*packets.Packet, bool) {
	p, ok := s.PendingPackets[packetID]
	return p, ok
}

func (s *Session) Remove(packetId uint16) {
	delete(s.PendingPackets, packetId)
}

func (s *Session) ClonePendingPackets() *Session {
	c := NewSession()

	for k, v := range s.PendingPackets {
		c.PendingPackets[k] = v
	}

	return c
}

func NewClient(conn net.Conn, broker *Broker) *Client {
	return &Client{
		Conn:    conn,
		Broker:  broker,
		Session: NewSession(),
	}
}

func (c *Client) GenerateClientID() {
	c.Propreties.ClientID = "PLACEHOLDER_FOR_GENERATED_CLIENT_ID"
}

func (c *Client) Send(buffer []byte) {
	_, err := c.Conn.Write(buffer)
	if err != nil {
		c.Close()
	}
}

func (c *Client) Close() {
	c.isClosed = true
	c.Conn.Close()
}

func (c *Client) IsClosed() bool {
	return c.isClosed
}

func (c *Client) ReadPackets() error {
	for {
		if c.IsClosed() {
			return nil
		}
		fixedHeader, err := packets.DecodeFixedHeader(c.Conn)
		if err != nil {
			return err
		}

		packet, err := packets.ParsePacket(fixedHeader, c.Conn)
		if err != nil {
			return err
		}

		c.HandlePacket(packet)
	}
}

func (c *Client) HandlePacket(packet *packets.Packet) {
	switch packet.FixedHeader.MessageType {
	case packets.CONNECT:
		c.HandleConnect(packet)
	case packets.DISCONNECT:
		c.HandleDisconnect(packet)
	case packets.SUBSCRIBE:
		c.HandleSubscribe(packet)
	case packets.UNSUBSCRIBE:
		c.HandleUnsubscribe(packet)
	case packets.PUBLISH:
		c.HandlePublish(packet)
	case packets.PUBACK:
		c.HandlePuback(packet)
	case packets.PUBREC:
		c.HandlePubrec(packet)
	case packets.PUBREL:
		c.HandlePubrel(packet)
	case packets.PUBCOMP:
		c.HandlePubcomp(packet)
	case packets.PINGREQ:
		c.HandlePingreq(packet)

	default:
		fmt.Println("TODO: implement handling of", packet.FixedHeader.MessageType)
	}
}

// Broker should not recive CONNECT packet here. If it does disconnect client
// TODO: check how it should handle in case for persistent sessions
func (c *Client) HandleConnect(packet *packets.Packet) {
	fmt.Println("Disconnecting client because of another CONNECT packet")
	c.Close()
}

func (c *Client) HandleDisconnect(packet *packets.Packet) {
	fmt.Println("Disconnecting client because of DISCONNECT packet")
	c.Close()
}

func (c *Client) HandlePingreq(packet *packets.Packet) {
	c.Send(packets.EncodePingresp())
}

func (c *Client) HandleSubscribe(packet *packets.Packet) {
	c.Broker.SubscribeClient(c, packet)
	suback := packet.EncodeSuback()
	c.Send(suback)
	fmt.Println("Subscribed to", packet.Subscriptions.OrderedSubscriptions)
}

func (c *Client) HandleUnsubscribe(packet *packets.Packet) {
	c.Broker.UnsubscribeClient(c, packet)
	unsuback := packet.EncodeUnsuback()
	c.Send(unsuback)
	fmt.Println("Unsubscribed from", packet.Subscriptions.OrderedSubscriptions)
}

func (c *Client) HandlePublish(packet *packets.Packet) {
	if packet.FixedHeader.Qos == 1 {
		fmt.Println("Sending PUBACK")
		c.Send(packet.EncodePuback())
	}

	if packet.FixedHeader.Qos == 2 {
		if _, ok := c.Session.Get(packet.PacketIdentifier); ok {
			fmt.Println("Packet identifier already in use")
			return
		}
		pubrec := packets.BuildResp(packet, packets.PUBREC)
		c.AddPendingPacket(pubrec)
		c.Send(pubrec.EncodeResp())
	}

	c.Broker.SendSubscribers(packet)
}

func (c *Client) HandlePuback(packet *packets.Packet) {
	fmt.Println("Received PUBACK")
	delete(c.Session.PendingPackets, packet.PacketIdentifier)
}

func (c *Client) HandlePubrec(packet *packets.Packet) {
	pubrel := packets.BuildResp(packet, packets.PUBREL)
	delete(c.Session.PendingPackets, packet.PacketIdentifier)
	c.AddPendingPacket(pubrel)
	c.Send(pubrel.EncodeResp())
}

func (c *Client) HandlePubrel(packet *packets.Packet) {
	if _, ok := c.Session.Get(packet.PacketIdentifier); !ok {
		fmt.Println("Packet identifier not found")
		return
	}

	pubcomp := packets.BuildResp(packet, packets.PUBCOMP)
	c.Session.Remove(packet.PacketIdentifier)
	c.Send(pubcomp.EncodeResp())
}

func (c *Client) HandlePubcomp(packet *packets.Packet) {
	delete(c.Session.PendingPackets, packet.PacketIdentifier)
	fmt.Println("Received PUBCOMP, qos transaction completed")
}

func (c *Client) SetClientPropreties(connectionOptions *packets.ConnectOptions) {
	c.Propreties = &Propreties{
		ProtocolLevel: connectionOptions.ProtocolLevel,
		ClientID:      connectionOptions.ClientID,
		Username:      connectionOptions.Username,
		Password:      connectionOptions.Password,
		WillTopic:     connectionOptions.WillTopic,
		WillMessage:   connectionOptions.WillMessage,
		WillRetain:    connectionOptions.WillRetain,
		WillQoS:       connectionOptions.WillQoS,
		CleanSession:  connectionOptions.CleanSession,
		Keepalive:     connectionOptions.Keepalive,
	}
}

func (c *Client) ValidateConnectionOptions() packets.Code {
	if c.Propreties.ProtocolLevel != ProtocolVersion {
		return packets.UNACCEPTABLE_PROTOCOL_VERSION
	}

	if c.Propreties.ClientID == "" || len(c.Propreties.ClientID) > 23 {
		return packets.IDENTIFIER_REJECTED
	}

	return packets.ACCEPTED
}

func (c *Client) AddPendingPacket(packet *packets.Packet) {
	c.Session.PendingPackets[packet.PacketIdentifier] = packet
}
