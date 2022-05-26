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

package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func Test_AuthOptions(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	stringToPtr := func(str string) *string {
		strPtr := &str
		return strPtr
	}

	input := `
auth:
  scheme: basic
  path_prefix: /login
  auth-upstream:
    host:
      hostname: example.com
      port: 80
`

	expected := &Auth{
		Scheme:     "basic",
		PathPrefix: stringToPtr("/login"),
		AuthUpstream: AuthUpstream{
			Host: &AuthUpstreamHost{
				Hostname: "example.com",
				Port:     80,
			},
		},
	}

	options := &SubOptions{}
	err := yaml.UnmarshalStrict([]byte(input), options)

	assert.NoError(err)
	assert.Equal(expected, options.Auth)
}
