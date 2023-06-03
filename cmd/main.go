package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/LanPavletic/nixMQ"
	"github.com/LanPavletic/nixMQ/listener"
)


func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	broker := nixmq.New()
	listener := listner.NewListener("127.0.0.1", "1883")
	broker.AddListener(listener)
	broker.Start()

	<-done
}
