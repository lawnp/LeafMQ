package main

import (
	"os"
	"os/signal"
	"syscall"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/LanPavletic/nixMQ"
	"github.com/LanPavletic/nixMQ/listener"
)


func main() {
	// this is for profiling memory usage
	// right now used for manually detecting memory leaks
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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
