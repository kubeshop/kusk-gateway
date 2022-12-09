package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	typedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

type websocketHandler struct {
	coreV1         typedCoreV1.CoreV1Interface
	writeWait      time.Duration
	pongWait       time.Duration
	pingPeriod     time.Duration
	maxMessageSize int64
}

const (
	defaultNamespaceParam = "kusk-system"
	defaultNameParam      = "kusk-gateway-envoy-fleet"
	defaultTailLineCount  = "1000"
)

func (h websocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	queryParams := r.URL.Query()
	namespace := queryParams.Get("namespace")
	if namespace == "" {
		namespace = defaultNamespaceParam
	}
	name := queryParams.Get("name")
	if name == "" {
		name = defaultNameParam
	}

	tailLineCount := queryParams.Get("tailLineCount")
	if tailLineCount == "" {
		tailLineCount = defaultTailLineCount
	}

	tailLineCountInt, err := strconv.Atoi(tailLineCount)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Errorf("invalid tailLineCount: %w", err).Error()))
		return
	}

	log.Println("client connected")
	stream, err := GetServiceContainerLogStream(
		r.Context(),
		namespace,
		name,
		"envoy",
		int64(tailLineCountInt),
		h.coreV1,
	)

	if err != nil {
		log.Printf("error getting log stream: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	defer stream.Close()

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	defer ws.Close()

	c := client{
		conn:           ws,
		logStream:      stream,
		writeWait:      h.writeWait,
		pongWait:       h.pongWait,
		pingPeriod:     h.pingPeriod,
		maxMessageSize: h.maxMessageSize,
	}

	stopCh := make(chan struct{})
	go c.readPump(r.Context(), stopCh)
	go c.writePump(r.Context(), stopCh)

	<-stopCh
	fmt.Println("request finished")
}
