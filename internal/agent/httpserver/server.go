/*
MIT License

Copyright (c) 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
// package server provides the Agent HTTP server, which is the service, configured with the Agent Management Service.
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
		Addr:         fmt.Sprintf("%s:%d", ServerHostname, ServerPort),
		Handler:      muxWithMiddlewares,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return server
}
