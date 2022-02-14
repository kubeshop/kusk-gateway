package main

import (
	"flag"
	"fmt"
	"time"

	"net/http"

	"github.com/kubeshop/kusk-gateway/internal/helper/server"
)

func main() {

	var (
		fleetID                                  string
		helperConfigurationManagerServiceAddress string
	)
	flag.StringVar(&helperConfigurationManagerServiceAddress, "helper-config-manager-service-address", "", "The address (hostname:port) of Kusk Gateway Helper Configuration Manager Service")
	flag.StringVar(&fleetID, "fleetID", "", "The Envoy Fleet ID this Helper server is deployed for.")
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle("/", server.NewHTTPHandler())
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", server.ServerHostname, server.ServerPort),
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("Starting the mocking server on ", server.Addr)
	server.ListenAndServe()
}
