/*
MIT License

# Copyright (c) 2022 Kubeshop

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
	"github.com/clbanning/mxj"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type mockedXmlRouteBuilder struct {
	*baseMockedRouteBuilder
}

func (m *mockedXmlRouteBuilder) BuildMockedRoute(args *BuildMockedRouteArgs) (*route.Route, error) {
	x, err := marshalXml(args.ExampleContent)
	if err != nil {
		return nil, err
	}

	return m.getRoute("xml", xmlPatternStr, string(x), args), nil
}

func marshalXml(content interface{}) ([]byte, error) {
	var xmlBytes []byte
	var err error
	if object, isObject := content.(map[string]interface{}); isObject {
		xml := mxj.Map(object)
		xmlBytes, err = xml.Xml()
	} else {
		xmlBytes, err = mxj.AnyXml(content, "root")
	}
	if err != nil {
		return nil, err
	}

	return xmlBytes, nil
}
