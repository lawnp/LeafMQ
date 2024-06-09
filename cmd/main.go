package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lawnp/nixMQ"
	"github.com/lawnp/nixMQ/listeners"
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

	tcp := listeners.NewTCP("127.0.0.1", "1883")
	broker.AddListener(tcp)
	broker.Start()

	<-done
}
