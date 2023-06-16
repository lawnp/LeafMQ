package nixmq

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/LanPavletic/nixMQ/listener"
	"github.com/LanPavletic/nixMQ/packets"
)

const (
	ProtocolVersion byte = 0x04
)

type Broker struct {
	listener []*listner.Listener // listener for incoming connections TODO: support multiple listeners
	clients map[string]*Client
	Subscriptions map[string][]*Client
	Log *log.Logger
}

func New() *Broker {
	return &Broker{
		clients: make(map[string]*Client),
		Log: initiateLog(),
		Subscriptions: make(map[string][]*Client),
	}
}

func initiateLog() *log.Logger{
	file, err := os.OpenFile(filepath.Join("logs", "broker.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		fmt.Println("Error opening log file:", err)
	}

	multipleOutput := io.MultiWriter(file, os.Stdout)
	return log.New(multipleOutput, "nixMQ: ", log.LstdFlags)
}

func (b *Broker) Start() {
	b.Log.Println("Starting broker")
	for _, l := range b.listener {
		go l.Serve(b.BindClient)
	}
}

func (b *Broker) AddListener(listner *listner.Listener) {
	b.listener = append(b.listener, listner)
}

func (b *Broker) AddClient(client *Client)  {
	if client.Propreties.ClientID == "" {
		b.Log.Println("ClientID is empty, creating new one")
		client.GenerateClientID()
	}

	b.clients[client.Propreties.ClientID] = client
}


func (b *Broker) BindClient(conn net.Conn) {	
	client := NewClient(conn, b)
	defer client.Close()

	connectPacket, err := b.ReadConnect(client)
	if err != nil {
		b.Log.Println("Error when trying to establish connection:", err)
		return
	}
	// set client properties
	client.SetClientPropreties(connectPacket.ConnectOptions)

	code := client.ValidateConnectionOptions()
	b.Log.Println("Validated connections options with:", code.Reason)
	b.sendConnack(client, code)
	
	if code != packets.ACCEPTED {
		return
	}

	b.AddClient(client)
	err = client.ReadPackets()

	if err != nil {
		b.Log.Println("Error Reading connections:", err)
		client.Close() // mybe not needed just in case
	}

	b.Log.Println("Client disconnected")
}

func (b *Broker) ReadConnect(client *Client) (*packets.Packet, error) {
	fixedHeader, err := packets.DecodeFixedHeader(client.Conn)
	if err != nil {
		return nil, err
	}

	if fixedHeader.MessageType != packets.CONNECT {
		return nil, fmt.Errorf("Expected CONNECT packet, got %v", fixedHeader.MessageType)
	}

	connect, err := packets.ParsePacket(fixedHeader, client.Conn)
	b.Log.Println("Received CONNECT packet")
	return connect, err
}

func (b *Broker) sendConnack(client *Client, code packets.Code) {
	sesionPresent := false
	connack := packets.NewConnack(code, sesionPresent)
	fmt.Println(connack)
	b.Log.Println("Sending connack")
	// TODO check if client is still connected
	client.Send(connack.Encode())
}

func (b *Broker) SendSubscribers(packet *packets.Packet) {
	for _, client := range b.Subscriptions[packet.PublishTopic] {
		maxQoS := client.Subscriptions[packet.PublishTopic]
		if maxQoS < packet.FixedHeader.Qos {
			packet.FixedHeader.Qos = maxQoS
			if maxQoS == 0 {
				packet.FixedHeader.RemainingLength -= 2
			}
		}
		buf := packet.EncodePublish()
		if packet.FixedHeader.Qos != 0 {
			client.AddPendingPacket(packet)
		}
		client.Send(buf)
	}
}
