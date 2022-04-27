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
// Callbacks are called by GRPC server on new events.
package manager

import (
	"context"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
)

type Callbacks struct {
	cacheMgr *cacheManager
	log      Logger
}

func (cb *Callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	// cb.log.Debug("stream", id, "open for", typ)
	return nil
}
func (cb *Callbacks) OnStreamClosed(id int64) {
	// cb.log.Debug("stream", id, "closed")
}

func (cb *Callbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {
	return nil
}

func (cb *Callbacks) OnDeltaStreamClosed(id int64) {
	// cb.log.Debug("delta stream", id, "closed")
}

func (cb *Callbacks) OnStreamRequest(id int64, r *discovery.DiscoveryRequest) error {
	if cb.cacheMgr.IsNodeExist(r.Node.Id) {
		return nil
	}
	if err := cb.cacheMgr.setNodeSnapshot(r.Node.Id, r.Node.Cluster); err != nil {
		// cb.log.Error(err)
		return err
	}
	return nil
}
func (cb *Callbacks) OnStreamResponse(context.Context, int64, *discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {
}
func (cb *Callbacks) OnStreamDeltaResponse(id int64, req *discovery.DeltaDiscoveryRequest, res *discovery.DeltaDiscoveryResponse) {
}
func (cb *Callbacks) OnStreamDeltaRequest(id int64, req *discovery.DeltaDiscoveryRequest) error {
	return nil
}
func (cb *Callbacks) OnFetchRequest(_ context.Context, req *discovery.DiscoveryRequest) error {
	return nil
}
func (cb *Callbacks) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {}
