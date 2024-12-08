package app

import (
	"context"
	"fmt"
	"net"

	cluster "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	logger "github.com/asishrs/proxyless-grpc-lb/common/pkg/logger"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
)

// const grpcMaxConcurrentStreams = 1000

func registerServices(grpcServer *grpc.Server, server xds.Server) {
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	endpoint.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	cluster.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	route.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	listener.RegisterListenerDiscoveryServiceServer(grpcServer, server)
}

// RunManagementServer starts an xDS server at the given port.
func RunManagementServer(ctx context.Context, server xds.Server, port uint, maxConcurrentStreams uint32) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(maxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Logger.Error("Failed to listen", zap.Error(err))
	}

	// register services
	registerServices(grpcServer, server)

	logger.Logger.Info("Management server listening", zap.Uint("port", port))
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logger.Logger.Error("Failed to start management server", zap.Error(err))
		}
	}()
	<-ctx.Done()

	grpcServer.GracefulStop()
}

// RunManagementGateway starts an HTTP gateway to an xDS server.
// func RunManagementGateway(ctx context.Context, srv xds.Server, port uint) {
// 	logger.Logger.Info("Gateway listening HTTP/1.1", zap.Uint("port", port))

// 	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: &xds.HTTPGateway{Server: srv}}
// 	go func() {
// 		if err := server.ListenAndServe(); err != nil {
// 			logger.Logger.Error("Failed to start gateway server", zap.Error(err))
// 		}
// 	}()

// }
