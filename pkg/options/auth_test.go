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
  scheme: basic
  path_prefix: /login
  auth-upstream:
    host:
      hostname: example.com
      port: 80
`,
			expected: &AuthOptions{
				// Scheme: "basic",
				Custom: &Custom{
					PathPrefix: StringToPtr("/login"),
					AuthUpstream: &AuthUpstream{
						Host: AuthUpstreamHost{
							Hostname: "example.com",
							Port:     80,
						},
					},
				},
			},
		},
		{
			input: `
auth:
  scheme: oauth2
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
				// Scheme: "oauth2",
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
  scheme: oauth2
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
				// Scheme: "oauth2",
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

			assert.NoError(err)
			assert.Equal(test.expected, options.Auth)
		})
	}
}

func Test_AuthOptions_Validate_OK(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	authOptions := &AuthOptions{
		// Scheme: "basic",
		Custom: &Custom{
			PathPrefix: StringToPtr("/login"),
			AuthUpstream: &AuthUpstream{
				Host: AuthUpstreamHost{
					Hostname: "example.com",
					Port:     80,
				},
			},
		},
	}

	options := &SubOptions{
		Auth: authOptions,
	}

	assert.NoError(options.Validate())
	// assert.Equal(expected, options.Auth)
}

func Test_AuthOptions_Validate_CloudEntity_OK(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	authOptions := &AuthOptions{
		// Scheme: "cloudentity",
		Custom: &Custom{
			PathPrefix: StringToPtr("/login"),
			AuthUpstream: &AuthUpstream{
				Host: AuthUpstreamHost{
					Hostname: "example.com",
					Port:     80,
				},
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
			AuthUpstream: &AuthUpstream{
				Host: AuthUpstreamHost{
					// Hostname: "example.com",
					Port: 80,
				},
			},
		},
	}

	options := &SubOptions{
		Auth: authOptions,
	}

	assert.EqualError(options.Validate(), "auth: (auth-upstream: (host: (hostname: cannot be blank; port: cannot be blank.).).).")
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
	ttt := "123"
	options := &SubOptions{
		Auth: &AuthOptions{
			OAuth2: &OAuth2{
				TokenEndpoint:         "!23",
				AuthorizationEndpoint: "321",
				Credentials: Credentials{
					ClientID:     "!23",
					ClientSecret: &ttt,
					ClientSecretRef: &ClientSecretRef{
						Name:      "!@3",
						Namespace: "412",
					},
					HmacSecret: "!@3",
				},
				RedirectURI:         "!@3",
				RedirectPathMatcher: "124",
				ForwardBearerToken:  true,
				AuthScopes:          []string{"1", "2"},
				Resources:           []string{"1", "2"},
			},
		},
	}
	y, _ := yaml.Marshal(options)
	fmt.Println(input)
	err := yaml.UnmarshalStrict([]byte(y), options)
	assert.NoError(err)

	assert.EqualError(options.Validate(), expected)
}
