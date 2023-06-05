package nixmq

import (
	"fmt"
	"net"

	"github.com/LanPavletic/nixMQ/listener"
	"github.com/LanPavletic/nixMQ/packets"
)

const (
	ProtocolVersion byte = 0x04
)

type Broker struct {
	listener []*listner.Listener // listener for incoming connections TODO: support multiple listeners
	clients map[string]*Client
}

func New() *Broker {
	return &Broker{
		clients: make(map[string]*Client),
	}
}

func (b *Broker) Start() {
	for _, l := range b.listener {
		go l.Serve(b.BindClient)
	}
}

func (b *Broker) AddListener(listner *listner.Listener) {
	b.listener = append(b.listener, listner)
}


func (b *Broker) BindClient(conn net.Conn) {	
	client := NewClient(conn)
	err := client.EstablishConnection()

	if err != nil {
		fmt.Println(err)
		client.Close()
		return
	}

	code := client.ValidateConnectionOptions()
	fmt.Printf("Validated connections options with reason: %s\n", code.Reason)
	b.sendConnack(client, code)
	for {}
}

func (b *Broker) sendConnack(client *Client, code packets.Code) {
	sesionPresent := false
	connack := packets.NewConnack(code, sesionPresent)
	fmt.Println(connack)
	client.Send(connack.Encode())
}
