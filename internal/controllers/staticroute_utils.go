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
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-logr/logr"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

const (
	pathRoute = "/"
)

// `path` cannot be root path of a service. Just a sanity check.
// See: https://github.com/kubeshop/kusk-gateway/issues/954.
func staticRouteCheckPaths(logger logr.Logger, opts *options.StaticOptions) error {
	for path := range opts.Paths {
		if path == pathRoute {
			err := fmt.Errorf("`StaticRoute` `path` cannot be root path of a service - path=%v", path)
			logger.Error(err, "invalid configuration detected in `paths`")
			return err
		}
	}

	return nil
}

// Call `staticRouteCheckPaths` first.
// See: https://github.com/kubeshop/kusk-gateway/issues/954.
func staticRouteAppendRootPath(logger logr.Logger, opts *options.StaticOptions) {
	logger.Info("`StaticRoute` staticRouteAppendRootPath before appending root", "opts", spew.Sprint(opts))

	// Append `/` path with all available `methods` pointing to `spec.upstream`.
	for _, method := range methods() {
		if opts.Paths[pathRoute] == nil || len(opts.Paths[pathRoute]) == 0 {
			opts.Paths[pathRoute] = make(map[options.HTTPMethod]*options.SubOptions)
		}

		// `upstream` should be defined at the `spec` level.
		opts.Paths[pathRoute][method] = &options.SubOptions{
			Upstream: &opts.Upstream,
		}

		logger.Info(
			"`StaticRoute` staticRouteAppendRootPath added new operation to `opts.Paths`",
			"method", method,
			// "path", path,
			"opts.Upstream", spew.Sprint(opts.Upstream),
			`opts.Paths[pathRoute][method]`, opts.Paths[pathRoute][method],
		)
	}

	logger.Info("staticRouteAppendRootPath after appending root", "opts.Paths", spew.Sprint(opts.Paths))
}

func methods() []options.HTTPMethod {
	return []options.HTTPMethod{
		options.HTTPMethod(http.MethodGet),
		options.HTTPMethod(http.MethodHead),
		options.HTTPMethod(http.MethodPost),
		options.HTTPMethod(http.MethodPut),
		options.HTTPMethod(http.MethodPatch),
		options.HTTPMethod(http.MethodDelete),
		options.HTTPMethod(http.MethodConnect),
		options.HTTPMethod(http.MethodOptions),
		options.HTTPMethod(http.MethodTrace),
	}
}
