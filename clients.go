package nixmq

type Clients struct {
	internal map[string]*Client
}

func NewClients() *Clients {
	return &Clients{
		make(map[string]*Client),
	}
}

func (c *Clients) Add(client *Client) {
	if client.Propreties.ClientID == "" {
		client.GenerateClientID()
	}

	c.internal[client.Propreties.ClientID] = client
}

func (c *Clients) Get(clientID string) (*Client, bool) {
	client, ok := c.internal[clientID]
	return client, ok
}

func (c *Clients) Remove(client *Client) {
	delete(c.internal, client.Propreties.ClientID)
}

func (c *Clients) GetAll() map[string]*Client {
	return c.internal
}

func (c *Clients) Len() int {
	return len(c.internal)
}