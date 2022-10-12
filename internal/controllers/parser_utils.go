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

package controllers

import (
	"fmt"
	"strings"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/getkin/kin-openapi/openapi3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kubeshop/kusk-gateway/internal/envoy/types"
	"github.com/kubeshop/kusk-gateway/pkg/options"
)

// extract Params returns a map mapping the name of a paths parameter to its schema
// where the schema elements we care about are its type and enum if its defined
func extractParams(parameters openapi3.Parameters) map[string]types.ParamSchema {
	params := map[string]types.ParamSchema{}

	for _, parameter := range parameters {
		// Prevent populating map with empty parameter names
		if parameter.Value != nil && parameter.Value.Name != "" {
			params[parameter.Value.Name] = types.ParamSchema{}

			// Extract the schema if it's not nil and assign the map value
			if parameter.Value.Schema != nil && parameter.Value.Schema.Value != nil {
				schemaValue := parameter.Value.Schema.Value

				// It is acceptable for Type and / or Enum to have their zero value
				// It means the user has not defined it, and we will construct the regex path accordingly
				params[fmt.Sprintf("{%s}", parameter.Value.Name)] = types.ParamSchema{
					Type: schemaValue.Type,
					Enum: schemaValue.Enum,
				}
			}
		}
	}

	return params
}

func generateRouteMatch(path string, method string, pathParameters map[string]types.ParamSchema, corsPolicy *route.CorsPolicy) *route.RouteMatch {
	headerMatcherConfig := []*route.HeaderMatcher{
		types.GetHeaderMatcherConfig([]string{strings.ToUpper(method)}, corsPolicy != nil),
	}

	routeMatcherBuilder := types.NewRouteMatcherBuilder(path, pathParameters)
	return routeMatcherBuilder.GetRouteMatcher(headerMatcherConfig)
}

func generateRedirect(redirectOpts *options.RedirectOptions) (*route.Route_Redirect, error) {
	if redirectOpts == nil {
		return nil, nil
	}

	builder := types.NewRouteRedirectBuilder().
		HostRedirect(redirectOpts.HostRedirect).
		PortRedirect(redirectOpts.PortRedirect).
		SchemeRedirect(redirectOpts.SchemeRedirect).
		PathRedirect(redirectOpts.PathRedirect).
		ResponseCode(redirectOpts.ResponseCode).
		StripQuery(redirectOpts.StripQuery)

	if redirectOpts.RewriteRegex != nil {
		builder = builder.RegexRedirect(redirectOpts.RewriteRegex.Pattern, redirectOpts.RewriteRegex.Substitution)
	}

	redirect, err := builder.ValidateAndReturn()
	if err != nil {
		return nil, err
	}

	return redirect, nil
}

func generateCORSPolicy(corsOpts *options.CORSOptions) (*route.CorsPolicy, error) {
	if corsOpts == nil {
		return nil, nil
	}

	return types.GenerateCORSPolicy(
		corsOpts.Origins,
		corsOpts.Methods,
		corsOpts.Headers,
		corsOpts.ExposeHeaders,
		corsOpts.MaxAge,
		corsOpts.Credentials,
	)
}

type HostPortPair struct {
	Host string
	Port uint32
}

func getUpstreamHost(upstreamOpts *options.UpstreamOptions) (*HostPortPair, error) {
	if upstreamOpts == nil {
		return nil, fmt.Errorf("cannot get upstream host and port from nil upstream options")
	}

	if upstreamOpts.Service != nil {
		return &HostPortPair{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local.", upstreamOpts.Service.Name, upstreamOpts.Service.Namespace),
			Port: upstreamOpts.Service.Port,
		}, nil
	}

	if upstreamOpts.Host != nil {
		return &HostPortPair{
			Host: upstreamOpts.Host.Hostname,
			Port: upstreamOpts.Host.Port,
		}, nil
	}

	return nil, fmt.Errorf("cannot get upstream host and port from upstream options")
}

// each cluster can be uniquely identified by dns name + port (i.e. canonical Host, which is hostname:port)
func generateClusterName(name string, port uint32) string {
	return fmt.Sprintf("%s-%d", name, port)
}

func generateMockID(path string, method string, operationID string) string {
	return fmt.Sprintf("%s-%s-%s", path, method, operationID)
}

func generateRateLimitStatPrefix(host, path, method, operationID string) string {
	return fmt.Sprintf("%s-%s-%s-%s", host, path, method, operationID)
}

func generateRoutePath(prefix, path string) string {
	if prefix == "" {
		return path
	}

	// Avoids path joins (removes // in e.g. /path//subpath, or //subpath)
	return fmt.Sprintf(`%s/%s`, strings.TrimSuffix(prefix, "/"), strings.TrimPrefix(path, "/"))
}

func generateRoute(
	clusterName string,
	corsPolicy *route.CorsPolicy,
	rewriteRegex *options.RewriteRegex,
	QoS *options.QoSOptions,
	websocket *bool,
) (*route.Route_Route, error) {

	var rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute
	if rewriteRegex != nil {
		rewritePathRegex = types.GenerateRewriteRegex(rewriteRegex.Pattern, rewriteRegex.Substitution)
	}

	var (
		requestTimeout, requestIdleTimeout int64  = 0, 0
		retries                            uint32 = 0
	)
	if QoS != nil {
		retries = QoS.Retries
		requestTimeout = int64(QoS.RequestTimeout)
		requestIdleTimeout = int64(QoS.IdleTimeout)
	}

	routeRoute := &route.Route_Route{
		Route: &route.RouteAction{
			ClusterSpecifier: &route.RouteAction_Cluster{
				Cluster: clusterName,
			},
		},
	}

	if corsPolicy != nil {
		routeRoute.Route.Cors = corsPolicy
	}
	if rewritePathRegex != nil {
		routeRoute.Route.RegexRewrite = rewritePathRegex
	}

	if requestTimeout != 0 {
		routeRoute.Route.Timeout = &durationpb.Duration{Seconds: requestTimeout}
	}
	if requestIdleTimeout != 0 {
		routeRoute.Route.IdleTimeout = &durationpb.Duration{Seconds: requestIdleTimeout}
	}

	if retries != 0 {
		routeRoute.Route.RetryPolicy = &route.RetryPolicy{
			RetryOn:    "5xx",
			NumRetries: &wrapperspb.UInt32Value{Value: retries},
		}
	}
	if websocket != nil && *websocket {
		routeRoute.Route.UpgradeConfigs = append(routeRoute.Route.UpgradeConfigs, &route.RouteAction_UpgradeConfig{UpgradeType: "websocket"})
	}
	if err := routeRoute.Route.ValidateAll(); err != nil {
		return nil, fmt.Errorf("incorrect Route Action: %w", err)
	}

	return routeRoute, nil
}

func mapRateLimitConf(rlOpt *options.RateLimitOptions, statPrefix string) *ratelimit.LocalRateLimit {
	var seconds int64
	switch rlOpt.Unit {
	case "second":
		seconds = 1
	case "minute":
		seconds = 60
	case "hour":
		seconds = 60 * 60
	}

	responseCode := rlOpt.ResponseCode
	if responseCode == 0 {
		// HTTP Status too many requests
		responseCode = 429
	}

	rl := &ratelimit.LocalRateLimit{
		StatPrefix: statPrefix,
		Status: &envoy_type_v3.HttpStatus{
			Code: envoy_type_v3.StatusCode(responseCode),
		},
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens: rlOpt.RequestsPerUnit,
			TokensPerFill: &wrapperspb.UInt32Value{
				Value: rlOpt.RequestsPerUnit,
			},
			FillInterval: &durationpb.Duration{
				Seconds: seconds,
			},
		},
		FilterEnabled: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enabled",
		},
		FilterEnforced: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enforced",
		},
		Stage:                                 0,
		LocalRateLimitPerDownstreamConnection: rlOpt.PerConnection,
	}

	return rl
}
