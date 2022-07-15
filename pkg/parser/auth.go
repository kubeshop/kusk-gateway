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

package parser

import (
	"fmt"

	envoy_config_filter_http_ext_authz_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	http "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/go-logr/logr"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/kubeshop/kusk-gateway/internal/envoy/config"
	"github.com/kubeshop/kusk-gateway/pkg/options"
)

var (
	ErrorAuthIsNil                = fmt.Errorf("parser.auth.ParseAuthOptions: `auth` is nil")
	ErrorMutuallyExclusiveOptions = fmt.Errorf("parser.auth.ParseAuthOptions: `auth.auth-upstream` and `auth.oauth2` are enabled but are mutually exclusive")
)

type ParseAuthArguments struct {
	EnvoyConfiguration           *config.EnvoyConfiguration
	HTTPConnectionManagerBuilder *config.HCMBuilder
}

func ParseAuthOptions(logger logr.Logger, finalOpts options.SubOptions, arguments ParseAuthArguments) error {
	if finalOpts.Auth == nil {
		return ErrorAuthIsNil
	}

	authUpstream := finalOpts.Auth.AuthUpstream
	oauth2 := finalOpts.Auth.OAuth2

	if authUpstream != nil && oauth2 != nil {
		return ErrorMutuallyExclusiveOptions
	}

	if authUpstream != nil {
		err := ParseAuthUpstreamOptions(authUpstream, arguments)
		if err != nil {
			return err
		}
	}

	if oauth2 != nil {
		err := ParseOAuth2Options(oauth2, arguments)
		if err != nil {
			return err
		}
	}

	logger.
		WithName("pkg/parser/auth.go:ParseAuthOptions").
		Info("added filter", "HTTPConnectionManager.HttpFilters", len(arguments.HTTPConnectionManagerBuilder.HTTPConnectionManager.HttpFilters))

	return nil
}

func ParseOAuth2Options(oauth2Options *options.OAuth2, arguments ParseAuthArguments) error {
	typedConfig, err := NewFilterHTTPOAuth2(oauth2Options, arguments)
	if err != nil {
		return err
	}

	filter := &http.HttpFilter{
		Name: "envoy.filters.http.oauth2",
		ConfigType: &http.HttpFilter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}

	return arguments.HTTPConnectionManagerBuilder.AddFilter(filter)
}

func ParseAuthUpstreamOptions(authUpstreamOptions *options.AuthUpstream, arguments ParseAuthArguments) error {
	upstreamServiceHost := authUpstreamOptions.Host.Hostname
	upstreamServicePort := authUpstreamOptions.Host.Port

	clusterName := GenerateClusterName(upstreamServiceHost, upstreamServicePort)

	if !arguments.EnvoyConfiguration.ClusterExist(clusterName) {
		arguments.EnvoyConfiguration.AddCluster(
			clusterName,
			upstreamServiceHost,
			upstreamServicePort,
		)
	}

	pathPrefix := ""
	if authUpstreamOptions.PathPrefix != nil {
		pathPrefix = *authUpstreamOptions.PathPrefix
	}

	typedConfig, err := NewFilterHTTPExternalAuthorization(
		upstreamServiceHost,
		upstreamServicePort,
		clusterName,
		pathPrefix,
	)
	if err != nil {
		return err
	}

	filter := &http.HttpFilter{
		Name: wellknown.HTTPExternalAuthorization,
		ConfigType: &http.HttpFilter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}

	return arguments.HTTPConnectionManagerBuilder.AddFilter(filter)
}

// RouteAuthzDisabled
// returns a per-route config to disable authorization.
func RouteAuthzDisabled() (*anypb.Any, error) {
	return anypb.New(
		&envoy_config_filter_http_ext_authz_v3.ExtAuthzPerRoute{
			Override: &envoy_config_filter_http_ext_authz_v3.ExtAuthzPerRoute_Disabled{
				Disabled: true,
			},
		},
	)
}
