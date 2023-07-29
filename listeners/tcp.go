package listeners

import (
	"net"
)

type TcpListener struct {
	address string
	port    string
}

func NewTCP(address string, port string) *TcpListener {
	return &TcpListener{address, port}
}

func (l *TcpListener) Serve(bind BindFn) error {
	ln, err := net.Listen("tcp", l.address+":"+l.port)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go bind(conn)
	}
}
