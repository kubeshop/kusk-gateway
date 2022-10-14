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

package options

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func Test_AuthOptions_UnmarshalStrict(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    string
		expected *AuthOptions
	}

	testCases := []testCase{
		{
			input: `
auth:
  custom:
    path_prefix: /login
    host:
       hostname: example.com
`,
			expected: &AuthOptions{
				Custom: &Custom{
					PathPrefix: StringToPtr("/login"),
					Host: AuthUpstreamHost{
						Hostname: "example.com",
						// Port:     80,
					},
				},
			},
		},
		{
			input: `
auth:
  oauth2:
    token_endpoint: https://oauth2.googleapis.com/token
    authorization_endpoint: https://accounts.google.com/o/oauth2/auth
    credentials:
      client_id: "client_id"
      client_secret: "client_secret"
      hmac_secret: "hmac_secret"
    redirect_uri: http://localhost
    redirect_path_matcher: /oauth2/callback
    signout_path: /oauth2/signout
    forward_bearer_token: true
    auth_scopes:
      - user
      - openid
    resources:
      - user
      - openid
`,
			expected: &AuthOptions{
				OAuth2: &OAuth2{
					TokenEndpoint:         "https://oauth2.googleapis.com/token",
					AuthorizationEndpoint: "https://accounts.google.com/o/oauth2/auth",
					Credentials: Credentials{
						ClientID:     "client_id",
						ClientSecret: StringToPtr("client_secret"),
						HmacSecret:   "hmac_secret",
						CookieNames: CookieNames{
							BearerToken:  "",
							OauthHMAC:    "",
							ExpiresOauth: "",
						},
					},
					RedirectURI:         "http://localhost",
					RedirectPathMatcher: "/oauth2/callback",
					SignoutPath:         "/oauth2/signout",
					ForwardBearerToken:  true,
					AuthScopes:          []string{"user", "openid"},
					Resources:           []string{"user", "openid"},
				},
			},
		},
		{
			input: `
auth:
  oauth2:
    token_endpoint: https://oauth2.googleapis.com/token
    authorization_endpoint: https://accounts.google.com/o/oauth2/auth
    credentials:
      client_id: "client_id"
      client_secret_ref:
        name: "some-secret-object-containing-client-id"
        namespace: "some-namespace"
      hmac_secret: hmac_secret
    redirect_uri: http://localhost
    redirect_path_matcher: /oauth2/callback
    signout_path: /oauth2/signout
    forward_bearer_token: true
    auth_scopes:
      - user
      - openid
    resources:
      - user
      - openid
`,
			expected: &AuthOptions{
				OAuth2: &OAuth2{
					TokenEndpoint:         "https://oauth2.googleapis.com/token",
					AuthorizationEndpoint: "https://accounts.google.com/o/oauth2/auth",
					Credentials: Credentials{
						ClientID: "client_id",
						ClientSecretRef: &ClientSecretRef{
							Name:      "some-secret-object-containing-client-id",
							Namespace: "some-namespace",
						},
						HmacSecret: "hmac_secret",
						CookieNames: CookieNames{
							BearerToken:  "",
							OauthHMAC:    "",
							ExpiresOauth: "",
						},
					},
					RedirectURI:         "http://localhost",
					RedirectPathMatcher: "/oauth2/callback",
					SignoutPath:         "/oauth2/signout",
					ForwardBearerToken:  true,
					AuthScopes:          []string{"user", "openid"},
					Resources:           []string{"user", "openid"},
				},
			},
		},
	}

	for index, testCase := range testCases {
		index, test := index, testCase
		name := fmt.Sprintf("%s-%d", t.Name(), index)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)
			options := &SubOptions{}
			err := yaml.UnmarshalStrict([]byte(test.input), options)
			jsn, _ := yaml.Marshal(test.expected)
			fmt.Println(string(jsn))
			assert.NoError(err)
			assert.Equal(test.expected, options.Auth)
		})
	}
}

func Test_AuthOptions_Validate_OK(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	authOptions := &AuthOptions{
		Custom: &Custom{
			PathPrefix: StringToPtr("/login"),
			Host: AuthUpstreamHost{
				Hostname: "example.com",
				Port:     80,
			},
		},
	}

	options := &SubOptions{
		Auth: authOptions,
	}

	assert.NoError(options.Validate())
}

func Test_AuthOptions_Validate_CloudEntity_OK(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	authOptions := &AuthOptions{
		Custom: &Custom{
			PathPrefix: StringToPtr("/login"),
			Host: AuthUpstreamHost{
				Hostname: "example.com",
				Port:     80,
			},
		},
	}

	options := &SubOptions{
		Auth: authOptions,
	}

	assert.NoError(options.Validate())
}

func Test_AuthOptions_Validate_Error(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	authOptions := &AuthOptions{
		Custom: &Custom{
			PathPrefix: StringToPtr("/login"),
			Host: AuthUpstreamHost{
				// Hostname: "example.com",
				Port: 80,
			},
		},
	}

	options := &SubOptions{
		Auth: authOptions,
	}

	assert.EqualError(options.Validate(), "auth: (custom: (host: (hostname: cannot be blank.).).).")
}

func Test_AuthOptions_OAuth2_Mutually_Exclusive_Client_Secret_Options(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	expected := "auth: (oauth2: (credentials: oauth2: You cannot specify both `client_secret_ref` and `client_secret`, the options are mutually exclusive.).)."
	input := `
auth:
  oauth2:
    token_endpoint: https://oauth2.googleapis.com/token
    authorization_endpoint: https://accounts.google.com/o/oauth2/auth
    credentials:
      client_id: "client_id"
      client_secret: "client_secret"
      client_secret_ref:
        name: some-secret-object-containing-client-id
        namespace: some-namespace
      hmac_secret: hmac_secret
    redirect_uri: http://localhost
    redirect_path_matcher: /oauth2/callback
    signout_path: /oauth2/signout
    forward_bearer_token: true
    auth_scopes:
      - user
      - openid
    resources:
      - user
      - openid
`

	options := &SubOptions{}
	err := yaml.Unmarshal([]byte(input), options)
	assert.NoError(err)

	assert.EqualError(options.Validate(), expected)
}
