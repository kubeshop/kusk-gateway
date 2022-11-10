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
	RoutePath = "/"
)

// `path` cannot be root path of a service.
// See: https://github.com/kubeshop/kusk-gateway/issues/954.
func staticRouteCheckPaths(logger logr.Logger, opts *options.StaticOptions) error {
	for path := range opts.Paths {
		if path == RoutePath {
			err := fmt.Errorf("`path` cannot be root path of a service - path=%v", path)
			logger.Error(err, "invalid configuration detected in `paths`")
			return err
		}
	}

	return nil
}

// Call `staticRouteCheckPaths` first.
// See: https://github.com/kubeshop/kusk-gateway/issues/954.
func staticRouteAppendRootPath(logger logr.Logger, opts *options.StaticOptions) {
	// TODO(MBana): Cleanup code.

	logger.Info("staticRouteAppendRootPath before appending root `opts.Paths`", "opts", spew.Sprint(opts))

	if opts.Paths == nil || len(opts.Paths) == 0 {
		opts.Paths = make(map[string]options.StaticOperationSubOptions)
	}

	var path options.StaticOperationSubOptions
	var ok bool
	if path, ok = opts.Paths[RoutePath]; !ok {
		opts.Paths[RoutePath] = options.StaticOperationSubOptions{}
	}

	// Append `/` path with all available `methods` pointing to `spec.upstream`.
	for _, method := range methods() {
		path = make(map[options.HTTPMethod]*options.SubOptions)
		if opts.Paths[RoutePath] == nil || len(opts.Paths[RoutePath]) == 0 {
			opts.Paths[RoutePath] = make(map[options.HTTPMethod]*options.SubOptions)
		}

		// `upstream` should be defined at the `spec` level.
		subOptions := &options.SubOptions{
			Upstream: opts.Upstream,
		}
		opts.Paths[RoutePath][method] = subOptions

		logger.Info(
			"staticRouteAppendRootPath added new operation to `opts.Paths`",
			"method", method,
			"path", path,
			"opts.Upstream", spew.Sprint(opts.Upstream),
			`opts.Paths["/"][method]`, opts.Paths[RoutePath][method],
		)
	}

	logger.Info("staticRouteAppendRootPath after appending root `opts.Paths`", "opts", spew.Sprint(opts))
}

func methods() []options.HTTPMethod {
	// TODO: Remove commented out methods, below.
	return []options.HTTPMethod{
		options.HTTPMethod(http.MethodGet),
		// options.HTTPMethod(http.MethodHead),
		// options.HTTPMethod(http.MethodPost),
		// options.HTTPMethod(http.MethodPut),
		// options.HTTPMethod(http.MethodPatch),
		// options.HTTPMethod(http.MethodDelete),
		// options.HTTPMethod(http.MethodConnect),
		// options.HTTPMethod(http.MethodOptions),
		// options.HTTPMethod(http.MethodTrace),
	}
}
