package listner

import (
	"net"
)

type BindFn func(net.Conn)

type Listener struct {
	address string
	port    string
}

func NewListener(address string, port string) *Listener {
	return &Listener{address, port}
}

func (l *Listener) Serve(bind BindFn) {
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

		go bind(conn)
	}
}
