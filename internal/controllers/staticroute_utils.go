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

	"github.com/davecgh/go-spew/spew"
	"github.com/go-logr/logr"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

// `path` cannot be root path of a service.
// See: https://github.com/kubeshop/kusk-gateway/issues/954.
func staticRouteCheckPaths(logger logr.Logger, opts *options.StaticOptions) error {

	for path := range opts.Paths {
		if path == "/" {
			err := fmt.Errorf("`path` cannot be root path of a service - path=%v", path)
			logger.Error(err, "invalid configuration detected in `paths`")
			return err
		}
	}

	return nil
}

//
// See: https://github.com/kubeshop/kusk-gateway/issues/954.
func staticRouteAppendRootPath(logger logr.Logger, opts *options.StaticOptions) {
	if len(opts.Paths) == 0 {
		logger.Info("empty `opts.Paths`", "opts.Paths", spew.Sprint(opts.Paths), "opts", spew.Sprint(opts))

		// path := opts.Paths["/"]
		// methods := []string{"GET"}
		// for _, method := range methods {
		// 	method := options.HTTPMethod(method)

		// 	subOptions := &options.SubOptions{}
		// 	path[method] = subOptions

		// 	logger.Info(
		// 		"added new operation to `opts.Paths`",
		// 		"method", method,
		// 		"path", path,
		// 		"subOptions", spew.Sprint(subOptions),
		// 	)
		// }
	} else {
		logger.Info("non-empty `opts.Paths`", spew.Sprint(opts.Paths), "opts", spew.Sprint(opts))
	}
}
