package nixmq

import (
	"net"
	"fmt"

	"github.com/LanPavletic/nixMQ/packets"
)

type Client struct {
	Propreties *Propreties
	Conn net.Conn
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

func NewClient(conn net.Conn) *Client {
	return &Client{
		Conn: conn,
	}
}

func (c *Client) GenerateClientID() {
	c.Propreties.ClientID = "PLACEHOLDER_FOR_GENERATED_CLIENT_ID"
}

func (c *Client) Send(buffer []byte) {
	c.Conn.Write(buffer)
}

func (c *Client) Close() {
	c.Conn.Close()
}

func (c *Client) ReadPackets() {}

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
