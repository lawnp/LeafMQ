package listner

import (
	"fmt"
	"net"

	"github.com/LanPavletic/nixMQ/packets"
)

type Listener struct {
	address string
	port    string
}

func NewListener(address string, port string) *Listener {
	return &Listener{address, port}
}

func (l *Listener) Serve() {
	ln, err := net.Listen("tcp", l.address+":"+l.port)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		go handleConnection(conn)
	}
}

// parse connect packet
// read options and create client
// send connack packet
func handleConnection(conn net.Conn) {
	// packet default size is 64 bytes
	// needs testing for different packet sizes
	packet := make([]byte, 64)
	n, err := conn.Read(packet)
	if err != nil {
		panic(err)
	}
	fmt.Println("Received:", n, "bytes")
	for i := 0; i < n; i++ {
		fmt.Printf("%08b, %x\n", packet[i], packet[i])
	}

	connectOptions := packets.ParseConnect(packet)
	fmt.Printf("%+v\n", connectOptions)
	
	defer conn.Close()
}
