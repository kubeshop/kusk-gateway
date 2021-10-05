// Example configuring testing endpoint with dynamic management

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kubeshop/kusk-gateway/envoy/config"
	"github.com/kubeshop/kusk-gateway/envoy/manager"
)

var (
	l    Logger
	port uint
)

func init() {
	l = Logger{}

	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 18000, "xDS management server port")
}

func main() {
	flag.Parse()
	mgr := manager.New(context.Background(), port, l)
	// TODO: signal support
	go mgr.Start()
	// Create envoy configuration and apply it
	fleetName := "fleet1"
	vhosts := []string{"example.org", "example.com"}
	envoyConfig := config.New(vhosts)
	upstreamService := config.UpstreamService{Name: "petstore", Port: 8080}
	clusterName := "petstore"
	// Backend cluster
	envoyConfig.AddCluster(clusterName, upstreamService)
	envoyConfig.AddRoute("findByStatus", "/api/v3/pet/findByStatus", "GET", clusterName, upstreamService.Name)
	snap, err := envoyConfig.GenerateSnapshot()
	if err != nil {
		l.Fatal(err)
	}
	if err := mgr.ApplyNewFleetSnapshot(fleetName, snap); err != nil {
		l.Error(err)
	}
	fmt.Printf("%v", snap)
	// Block indefinitelly allowing manager to serve configuration to Envoy
	// Might as well use wg or nil channel.
	for {
	}
}

type Logger struct {
}

// Log to stdout only if Debug is true.
func (logger Logger) Debugf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}
func (logger Logger) Debug(args ...interface{}) {
	log.Print(args...)
}

// Log to stdout only if Debug is true.
func (logger Logger) Infof(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

func (logger Logger) Info(args ...interface{}) {
	log.Print(args...)
}

// Log to stdout always.
func (logger Logger) Warnf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}
func (logger Logger) Warn(args ...interface{}) {
	log.Print(args...)
}

// Log to stdout always.
func (logger Logger) Errorf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

func (logger Logger) Error(args ...interface{}) {
	log.Print(args...)
}

func (logger Logger) Fatal(args ...interface{}) {
	log.Print(args...)
	os.Exit(1)
}
