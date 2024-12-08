package app

import (
	"context"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"sync"

	logger "github.com/asishrs/proxyless-grpc-lb/common/pkg/logger"
	d3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"go.uber.org/zap"
)

// Report type
func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	logger.Logger.Debug("cb.Report()  callbacks", zap.Any("Fetches", cb.Fetches), zap.Any("Requests", cb.Requests))
}

// OnStreamOpen type
func (cb *Callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	logger.Logger.Debug("OnStreamOpen", zap.Int64("id", id), zap.String("type", typ))
	return nil
}

// OnStreamClosed type
func (cb *Callbacks) OnStreamClosed(id int64, node *core.Node) {
	logger.Logger.Debug("OnStreamClosed", zap.Int64("id", id))
}

// OnStreamRequest type
func (cb *Callbacks) OnStreamRequest(id int64, req *d3.DiscoveryRequest) error {
	logger.Logger.Debug("OnStreamRequest", zap.Int64("id", id), zap.Any("Request", req))
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Requests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

// OnStreamResponse type
func (cb *Callbacks) OnStreamResponse(ctx context.Context, id int64, req *d3.DiscoveryRequest, resp *d3.DiscoveryResponse) {
	logger.Logger.Debug("OnStreamResponse", zap.Int64("id", id), zap.Any("Request", req), zap.Any("Response ", resp))
	cb.Report()
}

// OnFetchRequest type
func (cb *Callbacks) OnFetchRequest(ctx context.Context, req *d3.DiscoveryRequest) error {
	logger.Logger.Debug("OnFetchRequest", zap.Any("Request", req))
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Fetches++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnDeltaStreamOpen(ctx context.Context, id int64, typ string) error {
	logger.Logger.Debug("OnDeltaStreamOpen", zap.Int64("id", id), zap.String("type", typ))
	return nil

}

func (cb *Callbacks) OnDeltaStreamClosed(id int64, node *core.Node) {
	logger.Logger.Debug("OnDeltaStreamClosed", zap.Int64("id", id), zap.Any("Node", node))
}

func (cb *Callbacks) OnStreamDeltaRequest(id int64, req *d3.DeltaDiscoveryRequest) error {
	logger.Logger.Debug("OnStreamDeltaRequest", zap.Int64("id", id), zap.Any("Request", req))
	return nil
}

func (cb *Callbacks) OnStreamDeltaResponse(id int64, req *d3.DeltaDiscoveryRequest, resp *d3.DeltaDiscoveryResponse) {
	logger.Logger.Debug("OnStreamDeltaResponse", zap.Int64("id", id), zap.Any("Request", req), zap.Any("Response", resp))
}

// OnFetchResponse type
func (cb *Callbacks) OnFetchResponse(req *d3.DiscoveryRequest, resp *d3.DiscoveryResponse) {
	logger.Logger.Debug("OnFetchResponse", zap.Any("Request", req), zap.Any("Response", resp))
}

// Callbacks for XD Server
type Callbacks struct {
	Signal   chan struct{}
	Fetches  int
	Requests int
	mu       sync.Mutex
}
