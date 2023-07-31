package listeners

import (
	"net"
)

type TcpListener struct {
	address string
	port    string
	close   chan bool
}

func NewTCP(address string, port string) *TcpListener {
	return &TcpListener{address, port, make(chan bool)}
}

func (l *TcpListener) Serve(bind BindFn) error {
	ln, err := net.Listen("tcp", l.address+":"+l.port)
	if err != nil {
		return err
	}

	go func() {
		<-l.close
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go bind(conn)
	}
}

func (l *TcpListener) Close() {
	l.close <- true
}
