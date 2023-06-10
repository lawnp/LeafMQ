package nixmq

import (
	"fmt"
	"net"

	"github.com/LanPavletic/nixMQ/packets"
)

type Client struct {
	Propreties *Propreties
	Conn net.Conn
	isClosed bool
	Subscriptions map[string]byte
	Broker *Broker
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

func NewClient(conn net.Conn, broker *Broker) *Client {
	return &Client{
		Conn: conn,
		Subscriptions: make(map[string]byte),
		Broker: broker,
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

func (c *Client) addSubscription(topic string, qos byte) {
	c.Subscriptions[topic] = qos
	c.Broker.Subscriptions[topic] = append(c.Broker.Subscriptions[topic], c)
}

// this scales very poorly...
func (c *Client) removeSubscription(topic string) {
	delete(c.Subscriptions, topic)
	if clients, ok := c.Broker.Subscriptions[topic]; ok {
		for i, client := range clients {
			if client == c {
				c.Broker.Subscriptions[topic] = append(c.Broker.Subscriptions[topic][:i], c.Broker.Subscriptions[topic][i+1:]...)
				break
			}
		}
	}
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
	// Subscriptions.Subscriptions... change that in the future please.
	for t, q := range packet.Subscriptions.Subscriptions {
		c.addSubscription(t, q)
	}

	suback := packet.EncodeSuback()
	c.Send(suback)
	fmt.Println("Subscribed to", packet.Subscriptions.OrderedSubscriptions)
}

func (c *Client) HandleUnsubscribe(packet *packets.Packet) {
	for _, t := range packet.Subscriptions.OrderedSubscriptions {
		c.removeSubscription(t)
	}
	unsuback := packet.EncodeUnsuback()
	c.Send(unsuback)
	fmt.Println("Unsubscribed from", packet.Subscriptions.OrderedSubscriptions)
}

func (c *Client) HandlePublish(packet *packets.Packet) {
	// first implementing only for QoS 0
	buf := packet.EncodePublish()
	c.Broker.SendSubscribers(buf, packet.FixedHeader.Qos, packet.PublishTopic)
}

func (c *Client) SetClientPropreties(connectionOptions *packets.ConnectOptions){
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

	if !c.Propreties.CleanSession {
		fmt.Println("TODO: implement session persistence")
		return packets.SERVER_UNAVAILABLE
	}

	if c.Propreties.ClientID == "" || len(c.Propreties.ClientID) > 23 {
		return packets.IDENTIFIER_REJECTED
	}

	return packets.ACCEPTED
}
