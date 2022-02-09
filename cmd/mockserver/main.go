package main

import (
	"flag"
	"fmt"
	"time"

	"net/http"

	"github.com/kubeshop/kusk-gateway/internal/mocking/mockserver"
)

func main() {

	var (
		fleetID                                string
		managerMockConfigurationServiceAddress string
	)
	flag.StringVar(&managerMockConfigurationServiceAddress, "manager-mocking-config-service", "", "The address (hostname:port) of Kusk Gateway Mocking Configuration Service")
	flag.StringVar(&fleetID, "fleetID", "", "The Envoy Fleet ID this Mocking server is deployed for.")
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle("/", mockserver.NewMockHTTPHandler())
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", mockserver.ServerHostname, mockserver.ServerPort),
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("Starting the mocking server on ", server.Addr)
	server.ListenAndServe()
}
