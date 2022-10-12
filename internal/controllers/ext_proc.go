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
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extproc "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	// `WellknownExtProc`` - HTTP filter name for `ExternalProcessor`.
	WellknownExtProc = "envoy.filters.http.ext_proc"
)

func mapExternalProcessorConfig(headers []*envoy_config_core_v3.HeaderValue) *extproc.ExtProcPerRoute {
	// fetch validation service host and port once
	// TODO: fetch kusk gateway validator service dynamically
	const validatorURL string = "kusk-gateway-validator-service.kusk-system.svc.cluster.local:17000"

	proc := &extproc.ExtProcPerRoute{
		Override: &extproc.ExtProcPerRoute_Overrides{
			Overrides: &extproc.ExtProcOverrides{
				GrpcService: &envoy_config_core_v3.GrpcService{
					TargetSpecifier: &envoy_config_core_v3.GrpcService_GoogleGrpc_{
						GoogleGrpc: &envoy_config_core_v3.GrpcService_GoogleGrpc{
							TargetUri:  validatorURL,
							StatPrefix: "external_proc",
						},
					},
					InitialMetadata: headers,
					Timeout:         nil,
				},
				ProcessingMode: &extproc.ProcessingMode{
					RequestHeaderMode:   extproc.ProcessingMode_SEND,
					ResponseHeaderMode:  extproc.ProcessingMode_SKIP,
					RequestBodyMode:     extproc.ProcessingMode_BUFFERED,
					ResponseBodyMode:    extproc.ProcessingMode_NONE,
					RequestTrailerMode:  extproc.ProcessingMode_SKIP,
					ResponseTrailerMode: extproc.ProcessingMode_SKIP,
				},
			},
		},
	}

	return proc
}

func externalProcessorConfigDisabled() (*anypb.Any, error) {
	return anypb.New(
		&extproc.ExtProcPerRoute{
			Override: &extproc.ExtProcPerRoute_Disabled{Disabled: true},
		})

}
