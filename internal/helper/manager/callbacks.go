// Callbacks are called by GRPC server on new events.
package manager

// import (
// 	"context"

// )

// type Callbacks struct {
// 	cacheMgr *cacheManager
// 	log      Logger
// }

// func (cb *Callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
// 	// cb.log.Debug("stream", id, "open for", typ)
// 	return nil
// }
// func (cb *Callbacks) OnStreamClosed(id int64) {
// 	// cb.log.Debug("stream", id, "closed")
// }

// func (cb *Callbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {
// 	return nil
// }

// func (cb *Callbacks) OnDeltaStreamClosed(id int64) {
// 	// cb.log.Debug("delta stream", id, "closed")
// }

// func (cb *Callbacks) OnStreamRequest(id int64, r *discovery.DiscoveryRequest) error {
// 	if cb.cacheMgr.IsNodeExist(r.Node.Id) {
// 		return nil
// 	}
// 	if err := cb.cacheMgr.setNodeSnapshot(r.Node.Id, r.Node.Cluster); err != nil {
// 		// cb.log.Error(err)
// 		return err
// 	}
// 	return nil
// }
// func (cb *Callbacks) OnStreamResponse(context.Context, int64, *discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {
// }
// func (cb *Callbacks) OnStreamDeltaResponse(id int64, req *discovery.DeltaDiscoveryRequest, res *discovery.DeltaDiscoveryResponse) {
// }
// func (cb *Callbacks) OnStreamDeltaRequest(id int64, req *discovery.DeltaDiscoveryRequest) error {
// 	return nil
// }
// func (cb *Callbacks) OnFetchRequest(_ context.Context, req *discovery.DiscoveryRequest) error {
// 	return nil
// }
// func (cb *Callbacks) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {}
