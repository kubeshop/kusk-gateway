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

package auth

import (
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubeshop/kusk-gateway/internal/cloudentity"
	"github.com/kubeshop/kusk-gateway/internal/envoy/config"
	"github.com/kubeshop/kusk-gateway/pkg/options"
)

var (
	ErrorAuthIsNil                = fmt.Errorf("auth.ParseAuthOptions: `auth` is nil")
	ErrorMutuallyExclusiveOptions = fmt.Errorf("auth.ParseAuthOptions: `auth.auth-upstream` and `auth.oauth2` are enabled but are mutually exclusive")
)

type generateClusterNameFunc func( /*name*/ string /*port*/, uint32) string

type parseAuthOptionsArguments struct {
	Logger                       logr.Logger
	EnvoyConfiguration           *config.EnvoyConfiguration
	HTTPConnectionManagerBuilder *config.HCMBuilder
	Name                         string
	RoutePath                    string
	Method                       string
	CloudEntityBuilder           *cloudentity.Builder
	GenerateClusterName          generateClusterNameFunc
	KubernetesClient             client.Client
}

func NewParseAuthOptionsArguments(
	logger logr.Logger,
	envoyConfiguration *config.EnvoyConfiguration,
	httpConnectionManagerBuilder *config.HCMBuilder,
	name string, routePathstring,
	method string,
	cloudEntityBuilder *cloudentity.Builder,
	generateClusterName generateClusterNameFunc,
	kubernetesClient client.Client,
) *parseAuthOptionsArguments {
	return &parseAuthOptionsArguments{
		Logger:                       logger,
		EnvoyConfiguration:           envoyConfiguration,
		HTTPConnectionManagerBuilder: httpConnectionManagerBuilder,
		Name:                         name,
		RoutePath:                    routePathstring,
		Method:                       method,
		CloudEntityBuilder:           cloudEntityBuilder,
		GenerateClusterName:          generateClusterName,
		KubernetesClient:             kubernetesClient,
	}
}

func ParseAuthOptions(finalOpts options.SubOptions, args *parseAuthOptionsArguments) error {
	if finalOpts.Auth == nil {
		return ErrorAuthIsNil
	}

	authUpstream := finalOpts.Auth.AuthUpstream
	oauth2 := finalOpts.Auth.OAuth2

	if authUpstream != nil && oauth2 != nil {
		return ErrorMutuallyExclusiveOptions
	}

	if authUpstream != nil {
		scheme := finalOpts.Auth.Scheme
		err := ParseAuthUpstreamOptions(authUpstream, args, scheme)
		if err != nil {
			return err
		}
	}

	if oauth2 != nil {
		err := ParseOAuth2Options(oauth2, args)
		if err != nil {
			return err
		}
	}

	args.Logger.
		WithName("auth.ParseAuthOptions").
		Info("added filter", "HTTPConnectionManager.HttpFilters", len(args.HTTPConnectionManagerBuilder.HTTPConnectionManager.HttpFilters))

	return nil
}
