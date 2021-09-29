package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

type Callbacks struct {
	Signal            chan struct{}
	Debug             bool
	Fetches           int
	Requests          int
	DeltaRequests     int
	DeltaResponses    int
	Cache             cache.SnapshotCache
	ClustersSnapshots map[string]cache.Snapshot
	mu                sync.Mutex
}

func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	log.Printf("server callbacks fetches=%d requests=%d\n", cb.Fetches, cb.Requests)
}
func (cb *Callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	if cb.Debug {
		log.Printf("stream %d open for %s\n", id, typ)
	}
	return nil
}
func (cb *Callbacks) OnStreamClosed(id int64) {
	if cb.Debug {
		log.Printf("stream %d closed\n", id)
	}
}
func (cb *Callbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {
	if cb.Debug {
		log.Printf("delta stream %d open for %s\n", id, typ)
	}
	return nil
}
func (cb *Callbacks) OnDeltaStreamClosed(id int64) {
	if cb.Debug {
		log.Printf("delta stream %d closed\n", id)
	}
}
func (cb *Callbacks) OnStreamRequest(id int64, r *discovery.DiscoveryRequest) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Requests++

	// If the snapshot for nodeID from request is not found in the cache, get it from the cluster
	if _, err := cb.Cache.GetSnapshot(r.Node.Id); err != nil {
		snapshot, ok := cb.ClustersSnapshots[r.Node.Cluster]
		if !ok {
			err := fmt.Sprintf("No snapshots are registered for cluster %s", r.Node.Cluster)
			l.Errorf(err)
			return fmt.Errorf(err)
		}
		// Add the snapshot to the cache
		if err := cb.Cache.SetSnapshot(context.Background(), r.Node.Id, snapshot); err != nil {
			l.Errorf("snapshot error %q for %+v", err, snapshot)
			os.Exit(1)
		}
	}

	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}
func (cb *Callbacks) OnStreamResponse(context.Context, int64, *discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {
}
func (cb *Callbacks) OnStreamDeltaResponse(id int64, req *discovery.DeltaDiscoveryRequest, res *discovery.DeltaDiscoveryResponse) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaResponses++
}
func (cb *Callbacks) OnStreamDeltaRequest(id int64, req *discovery.DeltaDiscoveryRequest) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaRequests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}

	return nil
}
func (cb *Callbacks) OnFetchRequest(_ context.Context, req *discovery.DiscoveryRequest) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Fetches++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}
func (cb *Callbacks) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {}
