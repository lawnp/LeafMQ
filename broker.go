package nixmq

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	listner "github.com/LanPavletic/nixMQ/listener"
	"github.com/LanPavletic/nixMQ/packets"
)

const (
	ProtocolVersion byte = 0x04
)

type Broker struct {
	listener      []*listner.Listener // listener for incoming connections
	clients       *Clients
	Subscriptions *TopicTree
	Log           *log.Logger
}

func New() *Broker {
	return &Broker{
		clients:       NewClients(),
		Log:           initiateLog(),
		Subscriptions: NewTopicTree(),
	}
}

// initiateLog creates a log file if it doesn't already exist. If the log file creation fails,
// it will fallback to logging only to stdout. On successful creation, it will log messages
// both to stdout and the log file. The function returns a *log.Logger instance for logging purposes.
func initiateLog() *log.Logger {
	// Create logs directory if it doesn't exist
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		fmt.Println("Error creating logs directory:", err)
		return log.New(os.Stdout, "nixMQ: ", log.LstdFlags)
	}

	filePath := filepath.Join("logs", "broker.log")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return log.New(os.Stdout, "nixMQ: ", log.LstdFlags)
	}

	multiOut := io.MultiWriter(file, os.Stdout)
	return log.New(multiOut, "nixMQ: ", log.LstdFlags)
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

func (b *Broker) BindClient(conn net.Conn) {
	client := NewClient(conn, b)
	defer client.Close()

	connectPacket, err := b.ReadConnect(client)
	if err != nil {
		b.Log.Println("Error when trying to establish connection:", err)
		return
	}

	client.SetClientPropreties(connectPacket.ConnectOptions)

	code := client.ValidateConnectionOptions()
	sessionPresent := b.InheritSession(client)

	b.sendConnack(client, code, sessionPresent)

	if code != packets.ACCEPTED {
		return
	}

	b.clients.Add(client)
	err = client.ReadPackets()

	if err != nil {
		b.Log.Println("Error Reading connections:", err)
	}

	// need to delete client from clients

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

func (b *Broker) sendConnack(client *Client, code packets.Code, sesionPresent bool) {
	connack := packets.NewConnack(code, sesionPresent)
	fmt.Println(connack)
	b.Log.Println("Sending connack")
	// TODO check if client is still connected
	client.Send(connack.Encode())
}

func (b *Broker) SendSubscribers(packet *packets.Packet) {
	subscribers := b.Subscriptions.GetSubscribers(packet.PublishTopic)

	for client, maxQoS := range subscribers {

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

func (b *Broker) InheritSession(client *Client) bool {
	if oldClient, ok := b.clients.Get(client.Propreties.ClientID); ok {

		if client.Propreties.CleanSession {
			// need to implement session take over.
			return false
		}

		b.Log.Println("Client already exists, inheriting session")
		client.Session = oldClient.Session.ClonePendingPackets()
		
		for topic, qos := range oldClient.Session.Subscriptions.getAll() {
			b.Subscriptions.Remove(topic, oldClient)
			oldClient.Session.Subscriptions.remove(topic)

			b.Subscriptions.Add(topic, qos, client)
			client.Session.Subscriptions.add(topic, qos)
		}

		oldClient.Close()
		return true
	}

	return false
}

func (b *Broker) SubscribeClient(client *Client, packet *packets.Packet) {
	for topic, qos := range packet.Subscriptions.GetAll() {
		b.Subscriptions.Add(topic, qos, client)
		client.Session.Subscriptions.add(topic, qos)
		b.Log.Println("Client subscribed to topic:", topic)
	}
}

func (b *Broker) UnsubscribeClient(client *Client, packet *packets.Packet) {
	for topic := range packet.Subscriptions.GetAll() {
		b.Subscriptions.Remove(topic, client)
		client.Session.Subscriptions.remove(topic)
		b.Log.Println("Client subscribed to topic:", topic)
	}
}
