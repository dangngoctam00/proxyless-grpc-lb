package app

import (
	"context"
	"errors"
	"fmt"
	clusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	v3route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"strconv"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	lv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	ep "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	logger "github.com/asishrs/proxyless-grpc-lb/common/pkg/logger"
	"go.uber.org/zap"

	wrapperspb "github.com/golang/protobuf/ptypes/wrappers"

	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
)

type podEndPoint struct {
	IP     string
	Port   int32
	Weight int32
}

type podWeight struct {
	IP     string
	Weight int32
}

func getPodMetadata(clientset *kubernetes.Clientset, namespace string) map[string]podWeight {
	list, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	pods := list.Items
	ips := make(map[string]podWeight)
	for _, pod := range pods {
		labels := pod.Labels
		weight := 1
		if v, ok := labels["weight"]; ok {
			weight, _ = strconv.Atoi(v)
		}
		ip := pod.Status.PodIP
		ips[ip] = podWeight{ip, int32(weight)}
	}
	return ips
}

func getK8sEndPoints(serviceNames []string) (map[string][]podEndPoint, error) {
	k8sEndPoints := make(map[string][]podEndPoint)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	endPoints, err := clientset.CoreV1().Endpoints("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Logger.Error("Received error while trying to get EndPoints", zap.Error(err))
	}
	podWeights := getPodMetadata(clientset, "")
	for _, serviceName := range serviceNames {
		for _, endPoint := range endPoints.Items {
			name := endPoint.GetObjectMeta().GetName()
			//logger.Logger.Info("Label of endpoint: " + fmt.Sprintf("%+v", endPoint.GetObjectMeta()))
			if name == serviceName {
				var ips []string
				var ports []int32
				for _, subset := range endPoint.Subsets {
					for _, address := range subset.Addresses {
						ips = append(ips, address.IP)
					}
					for _, port := range subset.Ports {
						ports = append(ports, port.Port)
					}
				}
				logger.Logger.Debug("Endpoint", zap.String("name", name), zap.Any("IP Address", ips), zap.Any("Ports", ports))
				var podEndPoints []podEndPoint
				for _, port := range ports {
					for _, ip := range ips {
						weight := podWeights[ip]
						podEndPoints = append(podEndPoints, podEndPoint{ip, port, weight.Weight})
					}
				}
				k8sEndPoints[serviceName] = podEndPoints
			}
		}
	}
	return k8sEndPoints, nil
}

func clusterLoadAssignment(podEndPoints []podEndPoint, clusterName string, region string, zone string) []types.Resource {
	var lbs []*ep.LbEndpoint
	for _, podEndPoint := range podEndPoints {
		logger.Logger.Debug("Creating ENDPOINT", zap.String("host", podEndPoint.IP), zap.Int32("port", podEndPoint.Port))
		hst := &core.Address{Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  podEndPoint.IP,
				Protocol: core.SocketAddress_TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(podEndPoint.Port),
				},
			},
		}}

		lbs = append(lbs, &ep.LbEndpoint{
			HostIdentifier: &ep.LbEndpoint_Endpoint{
				Endpoint: &ep.Endpoint{
					Address: hst,
				}},
			HealthStatus:        core.HealthStatus_HEALTHY,
			LoadBalancingWeight: &wrapperspb.UInt32Value{Value: uint32(podEndPoint.Weight)},
		})
	}

	eds := []types.Resource{
		&ep.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*ep.LocalityLbEndpoints{{
				Locality: &core.Locality{
					Region: region,
					Zone:   zone,
				},
				Priority:            0,
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: uint32(1000)},
				LbEndpoints:         lbs,
			}},
		},
	}
	return eds
}

func createCluster(clusterName string) []types.Resource {
	logger.Logger.Debug("Creating CLUSTER", zap.String("name", clusterName))
	cls := []types.Resource{
		&clusterV3.Cluster{
			Name:                 clusterName,
			LbPolicy:             clusterV3.Cluster_ROUND_ROBIN,
			ClusterDiscoveryType: &clusterV3.Cluster_Type{Type: clusterV3.Cluster_EDS},
			EdsClusterConfig: &clusterV3.Cluster_EdsClusterConfig{
				EdsConfig: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{},
				},
			},
		},
	}
	return cls
}

func createVirtualHost(virtualHostName, listenerName, clusterName string) *v3route.VirtualHost {
	logger.Logger.Debug("Creating RDS", zap.String("host name", virtualHostName))
	vh := &v3route.VirtualHost{
		Name:    virtualHostName,
		Domains: []string{listenerName},

		Routes: []*v3route.Route{{
			Match: &v3route.RouteMatch{
				PathSpecifier: &v3route.RouteMatch_Prefix{
					Prefix: "",
				},
			},
			Action: &v3route.Route_Route{
				Route: &v3route.RouteAction{
					ClusterSpecifier: &v3route.RouteAction_Cluster{
						Cluster: clusterName,
					},
				},
			},
		}}}
	return vh

}

func createRoute(routeConfigName, virtualHostName, listenerName, clusterName string) []types.Resource {
	vh := createVirtualHost(virtualHostName, listenerName, clusterName)
	rds := []types.Resource{
		&v3route.RouteConfiguration{
			Name:         routeConfigName,
			VirtualHosts: []*v3route.VirtualHost{vh},
		},
	}
	return rds
}

func createListener(listenerName string, clusterName string, routeConfigName string) []types.Resource {
	logger.Logger.Debug("Creating LISTENER", zap.String("name", listenerName))
	hcRds := &hcm.HttpConnectionManager_Rds{
		Rds: &hcm.Rds{
			RouteConfigName: routeConfigName,
			ConfigSource: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{
					Ads: &core.AggregatedConfigSource{},
				},
			},
		},
	}

	routerConfig, _ := anypb.New(&router.Router{})

	filter := &hcm.HttpFilter{
		Name:       "envoy.filters.http.router",
		ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
	}
	manager := &hcm.HttpConnectionManager{
		CodecType:      hcm.HttpConnectionManager_AUTO,
		RouteSpecifier: hcRds,
		HttpFilters:    []*hcm.HttpFilter{filter},
	}

	pbst, err := ptypes.MarshalAny(manager)
	if err != nil {
		panic(err)
	}

	lds := []types.Resource{
		&listenerv3.Listener{
			Name: listenerName,
			ApiListener: &lv3.ApiListener{
				ApiListener: pbst,
			},
			Address: &core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Protocol: core.SocketAddress_TCP,
						Address:  "0.0.0.0",
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: 10000,
						},
					},
				},
			},
			FilterChains: []*listenerv3.FilterChain{{
				Filters: []*listenerv3.Filter{{
					Name: wellknown.HTTPConnectionManager,
					ConfigType: &listenerv3.Filter_TypedConfig{
						TypedConfig: pbst,
					},
				}},
			}},
		}}
	return lds
}

// GenerateSnapshot creates snapshot for each service
func GenerateSnapshot(services []string) (*cache.Snapshot, error) {
	k8sEndPoints, err := getK8sEndPoints(services)
	if err != nil {
		logger.Logger.Error("Error while trying to get EndPoints from k8s cluster", zap.Error(err))
		return nil, errors.New("Error while trying to get EndPoints from k8s cluster")
	}

	logger.Logger.Debug("K8s", zap.Any("EndPoints", k8sEndPoints))

	var eds []types.Resource
	var cds []types.Resource
	var rds []types.Resource
	var lds []types.Resource
	for service, podEndPoints := range k8sEndPoints {
		logger.Logger.Debug("Creating new XDS Entry", zap.String("service", service))
		clusterName := fmt.Sprintf("%s-cluster", service)
		eds = append(eds, clusterLoadAssignment(podEndPoints, clusterName, "my-region", "my-zone")...)
		cds = append(cds, createCluster(clusterName)...)
		rds = append(rds, createRoute(fmt.Sprintf("%s-route", service), fmt.Sprintf("%s-vhost", service), fmt.Sprintf("%s-listener", service), clusterName)...)
		lds = append(lds, createListener(fmt.Sprintf("%s-listener", service), clusterName, fmt.Sprintf("%s-route", service))...)
	}

	version := uuid.New()
	//logger.Logger.Debug("Creating Snapshot", zap.String("version", version.String()), zap.Any("EDS", eds), zap.Any("CDS", cds), zap.Any("RDS", rds), zap.Any("LDS", lds))
	logger.Logger.Debug("Creating Snapshot", zap.String("version", version.String()), zap.Any("EDS", eds), zap.Any("CDS", cds))

	resources := map[resource.Type][]types.Resource{
		resource.EndpointType: eds,
		resource.ClusterType:  cds,
		resource.RouteType:    rds,
		resource.ListenerType: lds,
	}

	snapshot, err := cache.NewSnapshot(version.String(), resources)

	if err != nil {
		logger.Logger.Error("Error while creating snapshot", zap.Error(err))
		return nil, errors.New("Error while creating snapshot")
	}

	if err := snapshot.Consistent(); err != nil {
		logger.Logger.Error("Snapshot inconsistency", zap.Any("snapshot", snapshot), zap.Error(err))
	}
	return snapshot, nil
}
