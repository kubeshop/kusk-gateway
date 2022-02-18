// package server provides the Helper HTTP server, which is the service, configured with the Helper Management Service.
package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	// Server Hostname and port are not configurable since the manager needs to know them for the Envoy cluster creation
	ServerHostname           string = "127.0.0.1"
	ServerPort               uint32 = 8090
	HeaderMockID                    = "X-Kusk-Mock-ID"
	HeaderMockResponseInsert        = "X-Kusk-Mocked"
)

type httpServer struct {
	mainHandler *mainHandler
	*http.Server
}

func NewHTTPServer(log *zap.Logger, mainHandler *mainHandler) *httpServer {
	server := &httpServer{}
	server.mainHandler = mainHandler

	mux := http.NewServeMux()
	muxWithMiddlewares := LoggerMiddleware(log, mux)
	mux.Handle("/", server.mainHandler)
	server.Server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", ServerHostname, ServerPort),
		Handler:        muxWithMiddlewares,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return server
}
