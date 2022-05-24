/*
MIT License

Copyright (c) 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package mocking

import (
	"encoding/json"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type mockedJsonRouteBuilder struct {
	baseMockedRouteBuilder
}

func (m mockedJsonRouteBuilder) BuildMockedRoute(args *BuildMockedRouteArgs) (*route.Route, error) {
	j, err := json.Marshal(args.ExampleContent)
	if err != nil {
		return nil, err
	}

	rt := m.getBaseRoute(args.RoutePath, args.Method)
	rt.Name = rt.Name + "-json"

	if args.RequireAcceptHeader {
		rt.Match.Headers = append(rt.Match.Headers, m.getAcceptHeaderMatcher(jsonPatternStr))
	}

	rt.Action = m.getAction(args.StatusCode, string(j))

	return rt, nil
}
