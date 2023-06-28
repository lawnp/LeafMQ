package nixmq

import (
	"fmt"
	"strings"

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
	return node.retained
}

func (t *TopicTree) Remove(topic string, client *Client) {
	topicLevels := splitTopic(topic)
	t.removeTopicRecursive(topicLevels, t.root, client)
}

func splitTopic(topic string) []string {
	return strings.Split(topic, "/")
}

func (t *TopicTree) getTopicNode(topicLevels []string, node *topicNode) *topicNode {
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
	if len(topicLevels) == 0 {
		node.subscribers.remove(client)
		return
	}

	topicLevel := topicLevels[0]
	childNode, ok := node.children[topicLevel]
	if !ok {
		fmt.Println("topic not found")
		return
	}
	t.removeTopicRecursive(topicLevels[1:], childNode, client)
}

func (t *TopicTree) GetSubscribers(topic string) map[*Client]byte {
	topicLevels := splitTopic(topic)

	node := t.root
	for _, topicLevel := range topicLevels {
		childNode, ok := node.children[topicLevel]
		if !ok {
			return nil
		}
		node = childNode
	}
	return node.subscribers.getAll()
}

func (t *TopicTree) GetAllTopics() []string {
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
	children    map[string]*topicNode // all child nodes
	subscribers *Subscribers          // array of client ids that are subscribed to this topic level
	retained	*packets.Packet       // retained message
}

func newTopicNode() *topicNode {
	return &topicNode{
		children:    make(map[string]*topicNode),
		subscribers: newSubscribers(),
	}
}

type Subscribers struct {
	clients map[*Client]byte
}

func newSubscribers() *Subscribers {
	return &Subscribers{
		clients: make(map[*Client]byte),
	}
}

func (s *Subscribers) add(client *Client, maxQoS byte) {
	s.clients[client] = maxQoS
}

func (s *Subscribers) remove(client *Client) {
	delete(s.clients, client)
}

func (s *Subscribers) getAll() map[*Client]byte {
	return s.clients
}

type Subscriptions struct {
	topics map[string]byte
}

func newSubscriptions() *Subscriptions {
	return &Subscriptions{
		topics: make(map[string]byte),
	}
}

func (s *Subscriptions) add(topic string, maxQoS byte) {
	s.topics[topic] = maxQoS
}

func (s *Subscriptions) remove(topic string) {
	delete(s.topics, topic)
}

func (s *Subscriptions) getAll() map[string]byte {
	return s.topics
}
