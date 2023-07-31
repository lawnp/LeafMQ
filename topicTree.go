package nixmq

import (
	"strings"
	"sync"
	"sync/atomic"

	"github.com/LanPavletic/nixMQ/packets"
)

type TopicTree struct {
	root *topicNode
}

func NewTopicTree() *TopicTree {
	return &TopicTree{
		root: newTopicNode(),
	}
}

func (t *TopicTree) Add(topic string, maxQoS byte, client *Client) *packets.Packet {
	topicLevels := splitTopic(topic)
	node := t.getTopicNode(topicLevels, t.root)
	node.subscribers.add(client, maxQoS)

	atomic.AddUint32(&client.Broker.Info.Subscriptions, 1)
	return node.retained
}

func (t *TopicTree) Remove(topic string, client *Client) {
	topicLevels := splitTopic(topic)
	t.removeTopicRecursive(topicLevels, t.root, client)

	// subtract 1 from the number of subscriptions
	// as per: https://pkg.go.dev/sync/atomic#AddUint32
	atomic.AddUint32(&client.Broker.Info.Subscriptions, ^uint32(0))
}

func splitTopic(topic string) []string {
	return strings.Split(topic, "/")
}

func (t *TopicTree) getTopicNode(topicLevels []string, node *topicNode) *topicNode {
	node.mu.Lock()
	defer node.mu.Unlock()

	if len(topicLevels) == 0 {
		return node
	}

	topicLevel := topicLevels[0]
	childNode, ok := node.children[topicLevel]
	if !ok {
		childNode = newTopicNode()
		node.children[topicLevel] = childNode
	}
	return t.getTopicNode(topicLevels[1:], childNode)
}

func (t *TopicTree) removeTopicRecursive(topicLevels []string, node *topicNode, client *Client) {
	node.mu.Lock()
	defer node.mu.Unlock()

	if len(topicLevels) == 0 {
		node.subscribers.remove(client)
		// if there are no subscribers, no retained message and no children we can remove the node
		if node.subscribers.Len() == 0 && node.retained == nil && len(node.children) == 0 {
			node = nil
		}
		return
	}

	topicLevel := topicLevels[0]
	childNode, ok := node.children[topicLevel]
	if !ok {
		return
	}
	t.removeTopicRecursive(topicLevels[1:], childNode, client)
}

func (t *TopicTree) GetSubscribers(topic string) *Subscribers {

	topicLevels := splitTopic(topic)
	subscribers := newSubscribers()

	node := t.root

	for i := 0; i < len(topicLevels); i++ {
		topicLevel := topicLevels[i]

		node.mu.RLock()
		defer node.mu.RUnlock()

		subscribers.addMultiLevelWildCard(node)

		if child, ok := node.children["+"]; ok {
			subscribers.addSingleLevelWildCard(child, topicLevels[i+1:])
		}

		childNode, ok := node.children[topicLevel]

		if !ok {
			break
		}

		node = childNode
	}

	subscribers.addMultiLevelWildCard(node)

	for client, qos := range node.subscribers.getAll() {
		subscribers.add(client, qos)
	}

	return subscribers
}

func (s *Subscribers) addMultiLevelWildCard(node *topicNode) {
	if child, ok := node.children["#"]; ok {
		for client, qos := range child.subscribers.getAll() {
			s.add(client, qos)
		}
	}
}

func (s *Subscribers) addSingleLevelWildCard(node *topicNode, topicLevels []string) {
	for _, topicLevel := range topicLevels {
		childNode, ok := node.children[topicLevel]

		if !ok {
			return
		}

		node = childNode
	}

	for client, qos := range node.subscribers.getAll() {
		s.add(client, qos)
	}
}

func (t *TopicTree) GetAllTopics() []string {
	t.root.mu.RLock()
	defer t.root.mu.RUnlock()
	topics := make([]string, 0)
	// tree traversal, append all topics to topics array
	getAllTopicsRecursive(t.root, "", &topics)
	return topics
}

func getAllTopicsRecursive(node *topicNode, topic string, topics *[]string) {
	if node.subscribers != nil && len(node.subscribers.getAll()) > 0 {
		*topics = append(*topics, topic)
	}

	for topicLevel, childNode := range node.children {
		childNode.mu.RLock()
		defer childNode.mu.RUnlock()
		if topic != "" {
			topicLevel = topic + "/" + topicLevel
		}
		getAllTopicsRecursive(childNode, topicLevel, topics)
	}
}

func (t *TopicTree) RemoveClientSubscriptions(client *Client) {
	for topic := range client.Session.Subscriptions.getAll() {
		t.Remove(topic, client)
	}
}

func (t *TopicTree) Retain(packet *packets.Packet) {
	topicLevels := splitTopic(packet.PublishTopic)
	t.getTopicNode(topicLevels, t.root).retained = packet
}

type topicNode struct {
	mu          sync.RWMutex
	children    map[string]*topicNode // all child nodes
	subscribers *Subscribers          // array of client ids that are subscribed to this topic level
	retained    *packets.Packet       // retained message
}

func newTopicNode() *topicNode {
	return &topicNode{
		children:    make(map[string]*topicNode),
		subscribers: newSubscribers(),
	}
}

type Subscribers struct {
	mu      sync.RWMutex
	clients map[*Client]byte
}

func newSubscribers() *Subscribers {
	return &Subscribers{
		clients: make(map[*Client]byte),
	}
}

func (s *Subscribers) add(client *Client, maxQoS byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client] = maxQoS
}

func (s *Subscribers) remove(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, client)
}

func (s *Subscribers) getAll() map[*Client]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients
}

func (s *Subscribers) Len() int {
	return len(s.clients)
}

type Subscriptions struct {
	mu     sync.RWMutex
	topics map[string]byte
}

func newSubscriptions() *Subscriptions {
	return &Subscriptions{
		topics: make(map[string]byte),
	}
}

func (s *Subscriptions) add(topic string, maxQoS byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.topics[topic] = maxQoS
}

func (s *Subscriptions) remove(topic string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.topics, topic)
}

func (s *Subscriptions) getAll() map[string]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.topics
}
