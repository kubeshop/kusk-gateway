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

package traffic

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"github.com/go-logr/logr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	HeaderXKuskWeightedCluster = "x-kusk-weighted-cluster"
)

func AddWeightedClusterToRoute(logger logr.Logger, routeRoute *envoy_config_route_v3.Route_Route, clusterName string, weight int) *envoy_config_route_v3.WeightedCluster {
	weightedClusters := routeRoute.Route.GetWeightedClusters()
	if weightedClusters == nil {
		weightedClusters = &envoy_config_route_v3.WeightedCluster{
			Clusters: []*envoy_config_route_v3.WeightedCluster_ClusterWeight{},
		}
	}

	if weightedClusters.Clusters == nil {
		weightedClusters.Clusters = []*envoy_config_route_v3.WeightedCluster_ClusterWeight{}
	}
	weightedClusters.Clusters = append(weightedClusters.Clusters, NewWeightedCluster(clusterName, uint32(weight)))

	logger.Info("adding weighted clusters", "weightedClusters.Clusters", weightedClusters.Clusters)

	return weightedClusters
}

func NewWeightedCluster(clusterName string, weight uint32) *envoy_config_route_v3.WeightedCluster_ClusterWeight {
	return &envoy_config_route_v3.WeightedCluster_ClusterWeight{
		Name:   clusterName,
		Weight: wrapperspb.UInt32(uint32(weight)),
		ResponseHeadersToAdd: []*envoy_config_core_v3.HeaderValueOption{
			{
				Header: &envoy_config_core_v3.HeaderValue{
					Key:   HeaderXKuskWeightedCluster,
					Value: clusterName,
				},
			},
		},
	}
}
