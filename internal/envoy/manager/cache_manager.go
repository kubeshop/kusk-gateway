// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package manager

import (
	"context"
	"fmt"
	"strings"
	"sync"

	cache_v3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/go-logr/logr"
)

// cacheManager provides cache and methods to update it with new configuration for Envoy fleet
// it is invisible for clients importing the package
type cacheManager struct {
	cache_v3.SnapshotCache
	fleetSnapshot map[string]*cache_v3.Snapshot // active snapshot per fleet
	mu            sync.RWMutex
	logger        logr.Logger
}

func NewCacheManager(snapshotCache cache_v3.SnapshotCache, logger logr.Logger) *cacheManager {
	return &cacheManager{
		SnapshotCache: snapshotCache,
		fleetSnapshot: map[string]*cache_v3.Snapshot{},
		mu:            sync.RWMutex{},
		logger:        logger.WithName("CacheManager"),
	}
}

func (cm *cacheManager) IsNodeExist(nodeID string) bool {
	if status := cm.GetStatusInfo(nodeID); status != nil {
		return true
	}

	return false
}

func (cm *cacheManager) getNodesWithCluster(cluster string) []string {
	var nodesIDs []string
	for _, nodeID := range cm.GetStatusKeys() {
		if cm.GetStatusInfo(nodeID).GetNode().Cluster == cluster {
			nodesIDs = append(nodesIDs, nodeID)
		}
	}

	return nodesIDs
}

// setNodeSnapshot sets new node snapshot from active fleet configuration snapshot
func (cm *cacheManager) setNodeSnapshot(nodeID string, fleet string) error {
	cm.mu.RLock()
	snapshot, ok := cm.fleetSnapshot[fleet]
	cm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no such %s Envoy fleet (cluster) configuration exist", fleet)
	}

	cm.logger.Info("setting new node snapshot", "nodeID", nodeID, "fleet", fleet)

	return cm.SetSnapshot(context.Background(), nodeID, snapshot)
}

// applyNewFleetSnapshot assigns active snapshot and updates all nodes with it
func (cm *cacheManager) applyNewFleetSnapshot(fleet string, newSnapshot *cache_v3.Snapshot) error {
	if err := newSnapshot.Consistent(); err != nil {
		return fmt.Errorf("inconsistent snapshot %v", newSnapshot)
	}

	cm.mu.Lock()
	cm.fleetSnapshot[fleet] = newSnapshot
	cm.mu.Unlock()

	errs := []error{}
	// Update caches for existing nodes with only this fleet
	for _, nodeID := range cm.getNodesWithCluster(fleet) {
		if err := cm.setNodeSnapshot(nodeID, fleet); err != nil {
			errs = append(errs, err)
		}
	}

	// Convert list of errors to strings.
	if len(errs) > 0 {
		errors := make([]string, len(errs))
		for i := 0; i < len(errs); i++ {
			errors[i] = errs[i].Error()
		}
		return fmt.Errorf(strings.Join(errors, "\n"))
	}

	cm.logger.Info("assigning active snapshot and updating all nodes", "fleet", fleet)

	return nil
}
