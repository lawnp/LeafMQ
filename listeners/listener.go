package listeners

import (
	"net"
)

type BindFn func(net.Conn)

type Listener interface {
	Serve(BindFn) error
	Close()
}
