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

// Callbacks are called by GRPC server on new events.
package manager

import (
	"context"

	envoy_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/go-logr/logr"
)

type Callbacks struct {
	cacheManager *cacheManager
	logger       logr.Logger
}

func NewCallbacks(cacheManager *cacheManager, logger logr.Logger) *Callbacks {
	return &Callbacks{
		cacheManager: cacheManager,
		logger:       logger.WithName("CacheManager"),
	}
}

func (c *Callbacks) OnStreamOpen(ctx context.Context, id int64, typeUrl string) error {
	c.logger.Info("OnStreamOpen", "id", id, "typeUrl", typeUrl)
	return nil
}
func (c *Callbacks) OnStreamClosed(id int64) {
	c.logger.Info("OnStreamClosed", "id", id)
}

func (c *Callbacks) OnDeltaStreamOpen(ctx context.Context, id int64, typeUrl string) error {
	c.logger.Info("OnDeltaStreamOpen", "id", id, "typeUrl", typeUrl)
	return nil
}

func (c *Callbacks) OnDeltaStreamClosed(id int64) {
	// `l.logger.V(1)` is effectively debug level.
	c.logger.Info("OnDeltaStreamClosed", "id", id)
}

func (c *Callbacks) OnStreamRequest(id int64, request *envoy_discovery_v3.DiscoveryRequest) error {
	c.logger.Info("OnStreamRequest", "id", id, "request.TypeUrl", request.TypeUrl)
	if c.cacheManager.IsNodeExist(request.Node.Id) {
		return nil
	}

	if err := c.cacheManager.setNodeSnapshot(request.Node.Id, request.Node.Cluster); err != nil {
		c.logger.Error(err, "OnStreamRequest", "id", id, "request.TypeUrl", request.TypeUrl)
		return err
	}

	return nil
}

func (c *Callbacks) OnStreamResponse(ctx context.Context, id int64, request *envoy_discovery_v3.DiscoveryRequest, response *envoy_discovery_v3.DiscoveryResponse) {
	c.logger.Info("OnStreamResponse", "id", id, "request.TypeUrl", request.TypeUrl, "response.TypeUrl", response.TypeUrl)
}

func (c *Callbacks) OnStreamDeltaResponse(id int64, request *envoy_discovery_v3.DeltaDiscoveryRequest, response *envoy_discovery_v3.DeltaDiscoveryResponse) {
	c.logger.Info("OnStreamDeltaResponse", "id", id, "request.TypeUrl", request.TypeUrl, "response.TypeUrl", response.TypeUrl)
}

func (c *Callbacks) OnStreamDeltaRequest(id int64, request *envoy_discovery_v3.DeltaDiscoveryRequest) error {
	c.logger.Info("OnStreamDeltaRequest", "id", id, "request.TypeUrl", request.TypeUrl)
	return nil
}

func (c *Callbacks) OnFetchRequest(ctx context.Context, request *envoy_discovery_v3.DiscoveryRequest) error {
	c.logger.Info("OnFetchRequest", "request.TypeUrl", request.TypeUrl)
	return nil
}

func (c *Callbacks) OnFetchResponse(request *envoy_discovery_v3.DiscoveryRequest, response *envoy_discovery_v3.DiscoveryResponse) {
	c.logger.Info("OnFetchResponse", "request.TypeUrl", request.TypeUrl, "response.TypeUrl", response.TypeUrl)
}
