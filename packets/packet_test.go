package packets

import (
	"reflect"
	"testing"
)

func TestCopy(t *testing.T) {
	fh := &FixedHeader{
		Dup:             false,
		Qos:             1,
		Retain:          true,
		MessageType:     1,
		RemainingLength: 42,
	}

	co := &ConnectOptions{
		ProtocolLevel: 4,
		ClientID:      "TestingClientId",
		Username:      "TestingUsername",
		Password:      "TestingPassword",
		WillTopic:     "WillTopic",
		WillMessage:   "WillMessage",
		WillRetain:    false,
		WillQoS:       1,
		CleanSession:  false,
		Keepalive:     30,
	}

	s := &Subscriptions{
		Subscriptions: map[string]byte{
			"topic/one": 0,
			"topic/two": 1,
		},
		OrderedSubscriptions: []string{"topic/one", "topic/two"},
	}

	originalPacket := &Packet{
		FixedHeader:      fh,
		ConnectOptions:   co,
		Subscriptions:    s,
		PublishTopic:     "PublishTopic",
		PacketIdentifier: 12345,
		Size:             64,
		Payload:          []byte{65, 66, 67},
	}

	copyPacket := originalPacket.Copy()

	// check if packets are the same
	if !reflect.DeepEqual(originalPacket, copyPacket) {
		t.Fatalf("Got %v, want %v", copyPacket, originalPacket)
	}

	// Change some values in the copied packet
	copyPacket.Payload = []byte{95, 96}
	copyPacket.ConnectOptions.ClientID = "ChangedClientId"
	copyPacket.FixedHeader.Dup = true
	copyPacket.Subscriptions.Subscriptions = map[string]byte{}

	if reflect.DeepEqual(originalPacket, copyPacket) {
		t.Fatal("Modified packet should not be the same as the original one", copyPacket)
	}

	// check if the original packet values stay the same
	if !reflect.DeepEqual(originalPacket.Payload, []byte{65, 66, 67}) ||
		!reflect.DeepEqual(originalPacket.Subscriptions.Subscriptions, map[string]byte{"topic/one": 0, "topic/two": 1}) ||
		originalPacket.ConnectOptions.ClientID != "TestingClientId" ||
		originalPacket.FixedHeader.Dup != false {

		t.Fatalf("Expacted original packet to stay the same but was changed: %v", originalPacket)

	}
}
