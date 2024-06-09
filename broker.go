package nixmq

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/lawnp/nixMQ/listeners"
	"github.com/lawnp/nixMQ/packets"
)

const (
	ProtocolVersion byte = 0x04
)

type Broker struct {
	listeners     []listeners.Listener // listeners for incoming connections
	clients       *Clients             // map of connected clients
	Subscriptions *TopicTree           // tree of topics and their subscribers
	Log           *log.Logger          // logger for logging messages
	Info          *Info                // Information about the broker (bytes sent, number of clients, etc.)
	Users         *Users               // map of users and their passwords (unencrypted in memory)
}

// creates a new broker instance
func New() *Broker {
	return &Broker{
		clients:       NewClients(),
		Log:           initiateLog(),
		Subscriptions: NewTopicTree(),
		Info:          &Info{},
		Users:         NewUsers(),
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
	b.Users.PopulateUsers() // testing purposes only

	errChan := make(chan error)

	for _, l := range b.listeners {
		go func(l listeners.Listener) {
			errChan <- l.Serve(b.BindClient)
		}(l)
	}

	go func() {
		for err := range errChan {
			b.Log.Println("Listener shut down:", err)
		}
	}()

	go b.handleCommands()
}

func (b *Broker) handleCommands() {
	for {
		var command string
		fmt.Scanln(&command)

		switch command {
		case "clients":
			b.Log.Println("Connected clients:", b.clients.Len())
			for clientId := range b.clients.GetAll() {
				b.Log.Println(clientId)
			}
		case "topics":
			b.Log.Println("Topics:")
			for _, topic := range b.Subscriptions.GetAllTopics() {
				b.Log.Println(topic)
			}
		case "info":
			b.DisplayInfo()
		}
	}
}

func (b *Broker) AddListener(listener listeners.Listener) {
	b.listeners = append(b.listeners, listener)
}

func (b *Broker) BindClient(conn net.Conn) {
	client := NewClient(conn, b)
	defer client.Close()

	connectPacket, err := b.ReadConnect(client)
	switch err.(type) {
	case nil:
	case *packets.ErrWrongProtocolLevel:
		b.sendConnack(client, packets.UNACCEPTABLE_PROTOCOL_VERSION, false)
		return
	case *packets.ErrWrongProtocolName:
		b.Log.Println("Packet format error:", err)
		return
	}

	client.SetClientProperties(connectPacket.ConnectOptions)

	code := client.ValidateConnectionOptions()
	sessionPresent := b.InheritSession(client)

	b.sendConnack(client, code, sessionPresent)

	if code != packets.ACCEPTED {
		return
	}

	b.clients.Add(client)
	client.RefreshKeepAlive()

	if sessionPresent {
		client.ResendPendingPackets()
	}

	err = client.ReadPackets()

	if err != nil {
		b.Log.Println("Error Reading connections:", err)
		client.SendWill()
	}
}

func (b *Broker) ReadConnect(client *Client) (*packets.Packet, error) {
	fixedHeader, err := packets.DecodeFixedHeader(client.Conn)
	if err != nil {
		return nil, err
	}

	if fixedHeader.MessageType != packets.CONNECT {
		return nil, fmt.Errorf("expected CONNECT packet, got %v", fixedHeader.MessageType)
	}

	connect, err := packets.ParsePacket(fixedHeader, client.Conn)
	b.Info.AddPacketReceived(connect)
	return connect, err
}

func (b *Broker) sendConnack(client *Client, code packets.Code, sesionPresent bool) {
	connack := packets.NewConnack(code, sesionPresent)
	// TODO check if client is still connected
	client.Send(connack.Encode())
}

// SendSubscribers sends a packet to all subscribers of a topic,
// adjusting the QoS and encoding the packet before sending it.
// If the packet has non-zero QoS, it adds it to the subscriber's
// pending packets list before sending.
func (b *Broker) SendSubscribers(packet *packets.Packet) {
	subscribers := b.Subscriptions.GetSubscribers(packet.PublishTopic)

	for client, maxQoS := range subscribers.getAll() {
		packet.SetRightQoS(maxQoS)
		buf := packet.EncodePublish()

		if packet.FixedHeader.Qos != 0 {
			client.AddPendingPacket(packet)
		}

		client.Send(buf)
	}
}

// InheritSession takes over the session of a client if a previous client with the same ClientID exists.
// If the client's CleanSession flag is true, the old client is closed and the session is not inherited.
// Returns true if session inheritance is successful, and false otherwise.
func (b *Broker) InheritSession(client *Client) bool {
	if oldClient, ok := b.clients.Get(client.Properties.ClientID); ok {

		// if clean session is true, we need to take over the session
		if client.Properties.CleanSession {
			b.CleanUp(oldClient)
			return false
		}

		client.Session = oldClient.Session.ClonePendingPackets()

		for topic, qos := range oldClient.Session.Subscriptions.getAll() {
			b.Subscriptions.Remove(topic, oldClient)
			oldClient.Session.Subscriptions.remove(topic)

			b.Subscriptions.Add(topic, qos, client)
			client.Session.Subscriptions.add(topic, qos)
		}

		b.CleanUp(oldClient)
		return true
	}

	return false
}

// SubscribeClient subscribes a client to the topics in the packet,
// adding it to the Broker's subscriptions and updating the client's session subscriptions.
// If there is a retained message for a topic, it sends a copy to the client with the adjusted QoS.
func (b *Broker) SubscribeClient(client *Client, packet *packets.Packet) {
	for topic, qos := range packet.Subscriptions.GetAll() {
		// In case topic filter was invalid, we don't need to do anything
		if qos != 0x80 {
			retained := b.Subscriptions.Add(topic, qos, client)
			client.Session.Subscriptions.add(topic, qos)

			if retained != nil {
				// needs to be copied because we might need to change the QoS
				retainedCopy := retained.Copy()
				retainedCopy.SetRightQoS(qos)
				client.Send(retainedCopy.EncodePublish())
			}
		}
	}
}

func (b *Broker) UnsubscribeClient(client *Client, packet *packets.Packet) {
	for _, topic := range packet.Subscriptions.GetOrdered() {
		b.Subscriptions.Remove(topic, client)
		client.Session.Subscriptions.remove(topic)
	}
}

func (b *Broker) CleanUp(client *Client) {
	b.clients.Remove(client)

	b.Subscriptions.RemoveClientSubscriptions(client)
	// I think this is needed to ensure that the client is garbage collected
	client.Broker = nil
}

func (b *Broker) DisplayInfo() {
	b.Log.Println("Broker Info:")
	b.Log.Println("Bytes received:", b.Info.BytesReceived)
	b.Log.Println("Bytes sent:", b.Info.BytesSent)
	b.Log.Println("Packets received:", b.Info.PacketsReceived)
	b.Log.Println("Packets sent:", b.Info.PacketsSent)
	b.Log.Println("Subscriptions:", b.Info.Subscriptions)
	b.Log.Println("Clients:", b.Info.Clients)
	b.Log.Println("Connected clients:", b.Info.ClientConnected)
	b.Log.Println("Disconnected clients:", b.Info.ClientDisconnected)
}

func (b *Broker) CloseAllListeners() {
	for _, listener := range b.listeners {
		listener.Close()
	}
}
func (b *Broker) Close() {
	b.CloseAllListeners()
	b.Log.Println("Closing broker")
}
