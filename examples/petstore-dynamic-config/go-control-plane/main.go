// Copyright 2020 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
package main

import (
	"context"
	"flag"
	"os"

	// "github.com/envoyproxy/go-control-plane/internal/
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

var (
	l    Logger
	port uint
)

const (
	cluster1 string = "envoy_cluster1"
	cluster2 string = "envoy_cluster2"
)

func init() {
	l = Logger{}

	flag.BoolVar(&l.Debug, "debug", false, "Enable xDS server debug logging")

	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 18000, "xDS management server port")
}

func main() {
	flag.Parse()

	// Create a cache
	c := cache.NewSnapshotCache(false, cache.IDHash{}, l)
	// Create the snapshot that we'll serve to Envoy for both clusters
	// NOTE: Technically they are the same, to make them different - new backend, new routes, etc
	snapshot1 := GenerateSnapshot()
	if err := snapshot1.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot1, err)
		os.Exit(1)
	}
	l.Debugf("will serve snapshot %+v", snapshot1)
	snapshot2 := GenerateSnapshot()
	if err := snapshot2.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot2, err)
		os.Exit(1)
	}
	l.Debugf("will serve snapshot %+v", snapshot2)
	// Run the xDS server
	ctx := context.Background()
	// Create active snapshots for each cluster
	cb := &Callbacks{
		Debug: l.Debug,
		Cache: c,
		ClustersSnapshots: map[string]cache.Snapshot{
			cluster1: snapshot1,
			cluster2: snapshot2,
		},
	}
	srv := server.NewServer(ctx, c, cb)
	RunServer(ctx, srv, port)
}
