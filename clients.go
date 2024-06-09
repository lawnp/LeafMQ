package nixmq

import (
	"sync"
	"sync/atomic"
)

type Clients struct {
	mu       sync.RWMutex
	internal map[string]*Client
}

func NewClients() *Clients {
	return &Clients{
		internal: make(map[string]*Client),
	}
}

func (c *Clients) Add(client *Client) {
	if client.Properties.ClientID == "" {
		client.GenerateClientID()
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.internal[client.Properties.ClientID] = client

	atomic.AddUint32(&client.Broker.Info.ClientConnected, 1)
	atomic.AddUint32(&client.Broker.Info.Clients, 1)
}

func (c *Clients) Get(clientID string) (*Client, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	client, ok := c.internal[clientID]
	return client, ok
}

func (c *Clients) Remove(client *Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.internal, client.Properties.ClientID)

	atomic.AddUint32(&client.Broker.Info.ClientDisconnected, ^uint32(0)) // --
	atomic.AddUint32(&client.Broker.Info.Clients, ^uint32(0))            // --
}

func (c *Clients) GetAll() map[string]*Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.internal
}

func (c *Clients) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.internal)
}
