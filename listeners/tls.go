package listeners

import (
	"crypto/tls"
)

type TlsListener struct {
	address string
	port    string
	config  *tls.Config
}

func NewTLS(address string, port string, config *tls.Config) *TlsListener {
	return &TlsListener{address, port, config}
}

func (l *TlsListener) Serve(bind BindFn) error {
	ln, err := tls.Listen("tcp", l.address+":"+l.port, l.config)
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
