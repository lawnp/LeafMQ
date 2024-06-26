package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	leafMQ "github.com/lawnp/leafMQ"
	"github.com/lawnp/leafMQ/listeners"
)

var (
	cpuProfile = flag.String("cpuprofile", "cpu.pprof", "file for cpu profile")
)

func main() {
	// Start the HTTP server for pprof
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	flag.Parse()

	f, err := os.Create("profiles/" + *cpuProfile)
	if err != nil {
		log.Fatal("could not create CPU profile file: ", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	// Signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	broker := leafMQ.New()

	tcp := listeners.NewTCP("127.0.0.1", "1883")
	broker.AddListener(tcp)
	broker.Start()

	select {
	case <-sigs:
	case <-time.After(30 * time.Second):
		break
	}
}
