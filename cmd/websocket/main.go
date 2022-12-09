package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	defaultWriteWaitSecs = 10

	// Time allowed to read the next pong message from the peer.
	defaultPongWaitSecs = 60

	// Maximum message size allowed from peer.
	defaultMaxMessageSize = 512
)

var (
	port           int
	writeWaitSecs  int
	pongWaitSecs   int
	maxMessageSize int64
)

func init() {
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.IntVar(&writeWaitSecs, "write-wait", defaultWriteWaitSecs, "time allowed to write a message to the peer")
	flag.IntVar(&pongWaitSecs, "pong-wait", defaultPongWaitSecs, "time allowed to read the next pong message from the peer")
	flag.Int64Var(&maxMessageSize, "max-message-size", defaultMaxMessageSize, "maximum message size allowed from peer")
	flag.Parse()
}

func main() {
	clientSet, err := NewClientSet()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	websocketHandler := websocketHandler{
		coreV1:    clientSet.CoreV1(),
		writeWait: time.Duration(writeWaitSecs) * time.Second,
		pongWait:  time.Duration(pongWaitSecs) * time.Second,
		// Send pings to peer with this period. Must be less than pongWait.
		pingPeriod:     (time.Duration(pongWaitSecs) * 9) / 10,
		maxMessageSize: maxMessageSize,
	}
	mux.Handle("/logs", websocketHandler)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	log.Println("starting server on :", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
