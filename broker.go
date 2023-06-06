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
	Log *log.Logger
}

func New() *Broker {
	return &Broker{
		clients: make(map[string]*Client),
		Log: initiateLog(),
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
	client := NewClient(conn)
	defer client.Close()

	err := client.EstablishConnection()
	if err != nil {
		fmt.Println(err)
		return
	}

	code := client.ValidateConnectionOptions()
	b.Log.Println("Validated connections options with:", code.Reason)
	b.sendConnack(client, code)
	
	if code != packets.ACCEPTED {
		return
	}

	b.AddClient(client)
	for {}
}

func (b *Broker) sendConnack(client *Client, code packets.Code) {
	sesionPresent := false
	connack := packets.NewConnack(code, sesionPresent)
	fmt.Println(connack)
	b.Log.Println("Sending connack")
	client.Send(connack.Encode())
}
