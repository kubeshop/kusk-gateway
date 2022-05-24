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
	"fmt"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type mockedTextRouteBuilder struct {
	baseMockedRouteBuilder
}

func (m mockedTextRouteBuilder) BuildMockedRoute(args *BuildMockedRouteArgs) (*route.Route, error) {
	text, err := getContentPlainTextReponse(args.ExampleContent)
	if err != nil {
		return nil, err
	}

	rt := m.getBaseRoute(args.RoutePath, args.Method)
	rt.Name = rt.Name + "-text"

	if args.RequireAcceptHeader {
		rt.Match.Headers = append(rt.Match.Headers, m.getAcceptHeaderMatcher(textPatternStr))
	}

	rt.Action = m.getAction(args.StatusCode, text)

	return rt, nil
}

func getContentPlainTextReponse(exampleContent interface{}) (string, error) {
	if bytes, ok := exampleContent.([]byte); ok {
		return string(bytes), nil
	} else if s, ok := exampleContent.(string); ok {
		return s, nil
	} else if s, ok := exampleContent.(fmt.Stringer); ok {
		// If it can't just be converted to string explicitly, call String method for that type if present.
		return s.String(), nil
	}

	return "", fmt.Errorf("cannot serialise %s into string", exampleContent)
}
