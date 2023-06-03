package nixmq

import (
	"github.com/LanPavletic/nixMQ/listener"
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
		go l.Serve()
	}
}

func (b *Broker) AddListener(listner *listner.Listener) {
	b.listener = append(b.listener, listner)
}
