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

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	SchemeBasic       = "basic"
	SchemeOAuth2      = "oauth2"
	SchemeCloudEntity = "cloudentity"
)

// AuthOptions example:
//
// x-kusk:
//
//	...
//	auth:
//	  scheme: basic
//	  auth-upstream:
//	    path_prefix: /login # optional
//	    host:
//	      hostname: example.com
//	      port: 80
type AuthOptions struct {
	// "basic", "cloudentity" "oauth2".
	// REQUIRED.
	Scheme string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	// OPTIONAL. TODO(MBana): Move to `AuthUpstream`.
	PathPrefix *string `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty"`
	// REQUIRED, if `scheme == basic`. Mutually exclusive with `OAuth2`.
	AuthUpstream *AuthUpstream `json:"auth-upstream,omitempty" yaml:"auth-upstream,omitempty"`
	// REQUIRED, if `scheme == oauth2`. Mutually exclusive with `AuthUpstream`.
	OAuth2 *OAuth2 `json:"oauth2,omitempty" yaml:"oauth2,omitempty"`
}

func (o AuthOptions) String() string {
	return ToCompactJSON(o)
}

func (o AuthOptions) Validate() error {
	err := validation.ValidateStruct(&o,
		validation.Field(&o.Scheme, validation.Required, validation.In(SchemeBasic, SchemeCloudEntity, SchemeOAuth2)),
	)
	if err != nil {
		return err
	}

	// `basic` or `cloudentity`
	if SchemeBasic == o.Scheme || SchemeCloudEntity == o.Scheme {
		return validation.ValidateStruct(&o,
			validation.Field(&o.AuthUpstream, validation.Required),
		)
	}

	// SchemeOAuth2
	return validation.ValidateStruct(&o,
		validation.Field(&o.OAuth2, validation.Required),
	)
}

type AuthUpstream struct {
	// REQUIRED.
	Host AuthUpstreamHost `json:"host,omitempty" yaml:"host,omitempty"`
	// OPTIONAL.
	PathPrefix *string `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty"`
}

func (o AuthUpstream) String() string {
	return ToCompactJSON(o)
}

func (o AuthUpstream) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Host, validation.Required),
	)
}

type AuthUpstreamHost struct {
	// REQUIRED.
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	// REQUIRED.
	Port uint32 `json:"port,omitempty" yaml:"port,omitempty"`
}

func (o AuthUpstreamHost) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Hostname, validation.Required, is.Host),
		// Do not attempt to validate with `is.Port`, otherwise `port: must be either a string or byte slice` error is returned.
		validation.Field(&o.Port, validation.Required),
	)
}

type OAuth2 struct {
	// Endpoint on the authorization server to retrieve the access token from.
	// REQUIRED.
	TokenEndpoint string `json:"token_endpoint,omitempty" yaml:"token_endpoint,omitempty"`
	// The endpoint redirect to for authorization in response to unauthorized requests.
	// REQUIRED.
	AuthorizationEndpoint string `json:"authorization_endpoint,omitempty" yaml:"authorization_endpoint,omitempty"`
	// Credentials used for OAuth.
	// REQUIRED.
	Credentials Credentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	// The redirect URI passed to the authorization endpoint. Supports header formatting tokens.
	// REQUIRED.
	RedirectURI string `json:"redirect_uri,omitempty" yaml:"redirect_uri,omitempty"`
	// Matching criteria used to determine whether a path appears to be the result of a redirect from the authorization server.
	// REQUIRED.
	RedirectPathMatcher string `json:"redirect_path_matcher,omitempty" yaml:"redirect_path_matcher,omitempty"`
	// The path to sign a user out, clearing their credential cookies.
	// REQUIRED.
	SignoutPath string `json:"signout_path,omitempty" yaml:"signout_path,omitempty"`
	// Forward the OAuth token as a Bearer to upstream web service.
	// When the authn server validates the client and returns an authorization token back to the OAuth filter, no matter what format that token is, if forward_bearer_token is set to true the filter will send over a cookie named BearerToken to the upstream. Additionally, the Authorization header will be populated with the same value.
	// REQUIRED.
	ForwardBearerToken bool `json:"forward_bearer_token,omitempty" yaml:"forward_bearer_token,omitempty"`
	// Optional list of OAuth scopes to be claimed in the authorization request.
	// If not specified, defaults to “user” scope. OAuth RFC https://tools.ietf.org/html/rfc6749#section-3.3.
	// OPTIONAL.
	AuthScopes []string `json:"auth_scopes,omitempty" yaml:"auth_scopes,omitempty"`
	// Optional resource parameter for authorization request RFC: https://tools.ietf.org/html/rfc8707.
	// OPTIONAL.
	Resources []string `json:"resources,omitempty" yaml:"resources,omitempty"`
}

func (o OAuth2) String() string {
	return ToCompactJSON(o)
}

func (o OAuth2) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.TokenEndpoint, validation.Required),
		validation.Field(&o.AuthorizationEndpoint, validation.Required),
		validation.Field(&o.Credentials, validation.Required),
		validation.Field(&o.RedirectURI, validation.Required),
		validation.Field(&o.RedirectPathMatcher, validation.Required),
		validation.Field(&o.SignoutPath, validation.Required),
		validation.Field(&o.ForwardBearerToken, validation.Required),
	)
}

type Credentials struct {
	// REQUIRED.
	ClientID string `json:"client_id,omitempty" yaml:"client_id,omitempty"`
	// REQUIRED, if `client_secret_ref` is not set, i.e., mutually exclusive with `client_secret_ref`.
	ClientSecret *string `json:"client_secret,omitempty" yaml:"client_secret,omitempty"`
	// REQUIRED, if `client_secret` is not set, i.e., mutually exclusive with `client_secret`.
	ClientSecretRef *ClientSecretRef `json:"client_secret_ref,omitempty" yaml:"client_secret_ref,omitempty"`
	// OPTIONAL.
	HmacSecret string `json:"hmac_secret,omitempty" yaml:"hmac_secret,omitempty"`
	// OPTIONAL.
	CookieNames CookieNames `json:"cookie_names,omitempty" yaml:"cookie_names,omitempty"`
}

func (o Credentials) String() string {
	return ToCompactJSON(o)
}

func (o Credentials) Validate() error {
	// You cannot specify both `client_secret_ref` and `client_secret`. An error is generated if both are specified.
	if o.ClientSecret != nil && o.ClientSecretRef != nil {
		return fmt.Errorf("oauth2: You cannot specify both `client_secret_ref` and `client_secret`, the options are mutually exclusive")
	}

	if o.ClientSecret != nil {
		return validation.ValidateStruct(&o,
			validation.Field(&o.ClientID, validation.Required),
			validation.Field(&o.ClientSecret, validation.Required),
		)
	}

	// o.ClientSecretRef != nil
	return validation.ValidateStruct(&o,
		validation.Field(&o.ClientID, validation.Required),
		validation.Field(&o.ClientSecretRef, validation.Required),
	)
}

type ClientSecretRef struct {
	// REQUIRED.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// REQUIRED.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

func (o ClientSecretRef) String() string {
	return ToCompactJSON(o)
}

func (o ClientSecretRef) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Name, validation.Required),
		validation.Field(&o.Namespace, validation.Required),
	)
}

// CookieNames - By default, OAuth2 filter sets some cookies with the following names: BearerToken, OauthHMAC, and OauthExpires. These cookie names can be customized by setting cookie_names.
type CookieNames struct {
	// Defaults to BearerToken.
	BearerToken string `json:"bearer_token,omitempty" yaml:"bearer_token,omitempty"`
	// Defaults to OauthHMAC.
	OauthHMAC string `json:"oauth_hmac,omitempty" yaml:"oauth_hmac,omitempty"`
	// Defaults to OauthExpires.
	ExpiresOauth string `json:"oauth_expires,omitempty" yaml:"oauth_expires,omitempty"`
}

func (o CookieNames) Validate() error {
	return nil
}
