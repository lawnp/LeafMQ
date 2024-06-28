package nixmq

import (
	"sync/atomic"

	"github.com/lawnp/leafMQ/packets"
)

type Info struct {
	BytesReceived      uint64
	BytesSent          uint64
	PacketsSent        uint64
	PacketsReceived    uint64
	Subscriptions      uint32
	Clients            uint32
	ClientDisconnected uint32
	ClientConnected    uint32
}

func (i *Info) AddPacketReceived(packet *packets.Packet) {
	atomic.AddUint64(&i.PacketsReceived, 1)
	atomic.AddUint64(&i.BytesReceived, uint64(packet.Size))
}

func (i *Info) AddPacketSent(n int64) {
	atomic.AddUint64(&i.PacketsSent, 1)
	atomic.AddUint64(&i.BytesSent, uint64(n))
}
