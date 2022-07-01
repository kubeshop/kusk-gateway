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

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"

	// xds_core_v3 "github.com/cncf/xds/go/xds/core/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_extensions_filter_http_oauth2_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/oauth2/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

func NewFilterHTTPOAuth2(oauth2Options *options.OAuth2, args *parseAuthOptionsArguments) (*anypb.Any, error) {
	// Example Input: "https://kubeshop-kusk-gateway-oauth2.eu.auth0.com/oauth/token"
	// Example Output from url.Hostname(): "kubeshop-kusk-gateway-oauth2.eu.auth0.com"
	// Example Output from url.Port(): ""
	url, err := url.Parse(oauth2Options.TokenEndpoint)
	if err != nil {
		err = fmt.Errorf("auth.NewFilterHTTPOAuth2: could not determine upstreamServiceHost from oauth2.token_endpoint=%q, %w", oauth2Options.TokenEndpoint, err)
		args.Logger.Error(err, "oauth2.token_endpoint contains an invalid url, failed to parse url")
		return nil, err
	}

	// // cluster := "kubeshop-kusk-gateway-oauth2.eu.auth0.com"
	// cluster := "eu.auth0.com"
	// sni := "eu.auth0.com"
	// // sni := "kubeshop-kusk-gateway-oauth2.eu.auth0.com"
	// // upstreamServiceHost := "kubeshop-kusk-gateway-oauth2.eu.auth0.com"
	// upstreamServiceHost := "eu.auth0.com"

	cluster := url.Hostname()
	upstreamServiceHost := url.Hostname()
	upstreamServicePort := uint32(443)
	if "" != url.Port() {
		port, err := strconv.ParseUint(url.Port(), 10, 32)
		if err != nil {
			err = fmt.Errorf("auth.NewFilterHTTPOAuth2: could not convert port=%q to int oauth2.token_endpoint=%q, %w", url.Port(), oauth2Options.TokenEndpoint, err)
			args.Logger.Error(err, "oauth2.token_endpoint contains an invalid port, failed to parse url")
			return nil, err
		}
		upstreamServicePort = uint32(port)
	}

	if !args.EnvoyConfiguration.ClusterExist(cluster) {
		err := args.EnvoyConfiguration.AddClusterWithTLS(cluster, upstreamServiceHost, upstreamServicePort)
		if err != nil {
			return nil, fmt.Errorf("auth.NewFilterHTTPOAuth2: failed on `arguments.EnvoyConfiguration.AddClusterWithTLS`, err=%w", err)
		}
		// args.EnvoyConfiguration.AddCluster(cluster, upstreamServiceHost, upstreamServicePort)
	} else {
		fmt.Println("----------------")
		fmt.Printf("auth.NewFilterHTTPOAuth2\n")
		fmt.Println("")
		fmt.Printf("cluster=%q exists\n", cluster)
		fmt.Printf("upstreamServiceHost=%q\n", upstreamServiceHost)
		fmt.Printf("upstreamServicePort=%v\n", upstreamServicePort)
		fmt.Printf("cluster=%q\n", cluster)
		// fmt.Printf("sni=%q\n", sni)
		fmt.Println("----------------")
	}

	httpUpstreamType := &envoy_config_core_v3.HttpUri_Cluster{
		Cluster: cluster,
	}
	tokenEndpoint := &envoy_config_core_v3.HttpUri{
		// Uri:              "https://kubeshop-kusk-gateway-oauth2.eu.auth0.com/oauth/token",
		Uri:              oauth2Options.TokenEndpoint,
		HttpUpstreamType: httpUpstreamType,
		Timeout:          TimeoutDefault(),
	}
	authorizationEndpoint := oauth2Options.AuthorizationEndpoint

	tokenSecret := &envoy_extensions_transport_sockets_tls_v3.SdsSecretConfig{
		Name: "token",
		// SdsConfig: &envoy_config_core_v3.ConfigSource{
		// 	Authorities:         authorities,
		// 	ResourceApiVersion:  ResourceApiVersion,
		// 	InitialFetchTimeout: TimeoutDefault(),
		// 	ConfigSourceSpecifier: &envoy_config_core_v3.ConfigSource_PathConfigSource{
		// 		PathConfigSource: &envoy_config_core_v3.PathConfigSource{
		// 			Path: "/etc/envoy/envoy.yaml",
		// 			WatchedDirectory: &envoy_config_core_v3.WatchedDirectory{
		// 				Path: "/etc/envoy",
		// 			},
		// 		},
		// 	},
		// 	// 	ConfigSourceSpecifier: &envoy_config_core_v3.ConfigSource_Path{
		// 	// 		Path: "/etc/envoy/token-secret.yaml.yaml",
		// 	// 	},
		// },
	}

	tokenFormation := &envoy_extensions_filter_http_oauth2_v3.OAuth2Credentials_HmacSecret{
		HmacSecret: &envoy_extensions_transport_sockets_tls_v3.SdsSecretConfig{
			Name: "hmac",
			// SdsConfig: &envoy_config_core_v3.ConfigSource{
			// 	Authorities:         authorities,
			// 	ResourceApiVersion:  ResourceApiVersion,
			// 	InitialFetchTimeout: TimeoutDefault(),
			// 	ConfigSourceSpecifier: &envoy_config_core_v3.ConfigSource_PathConfigSource{
			// 		PathConfigSource: &envoy_config_core_v3.PathConfigSource{
			// 			Path: "/etc/envoy/envoy.yaml",
			// 			WatchedDirectory: &envoy_config_core_v3.WatchedDirectory{
			// 				Path: "/etc/envoy",
			// 			},
			// 		},
			// 	},
			// 	// 	ConfigSourceSpecifier: &envoy_config_core_v3.ConfigSource_Path{
			// 	// 		Path: "/etc/envoy/hmac-secret.yaml",
			// 	// 	},
			// },
		},
	}
	credentials := &envoy_extensions_filter_http_oauth2_v3.OAuth2Credentials{
		// The client_id to be used in the authorize calls. This value will be URL encoded when sent to the OAuth server.
		ClientId: oauth2Options.Credentials.ClientID,
		// The secret used to retrieve the access token. This value will be URL encoded when sent to the OAuth server.
		TokenSecret: tokenSecret,
		// Configures how the secret token should be created.
		//
		// Types that are assignable to TokenFormation:
		//	*OAuth2Credentials_HmacSecret
		TokenFormation: tokenFormation,
	}

	// Becomes: "http://192.168.49.2/oauth2/callback"
	redirectUri := fmt.Sprintf("%s://%s%s", "%REQ(x-forwarded-proto)%", "%REQ(:authority)%", oauth2Options.RedirectURI)
	redirectPathMatcher := PathMatcherExact(oauth2Options.RedirectPathMatcher, false)
	signoutPath := PathMatcherExact(oauth2Options.SignoutPath, false)
	forwardBearerToken := oauth2Options.ForwardBearerToken
	// passThroughMatcher := []*envoy_config_route_v3.HeaderMatcher{
	// 	{},
	// }
	authScopes := oauth2Options.AuthScopes
	resources := oauth2Options.Resources

	config := &envoy_extensions_filter_http_oauth2_v3.OAuth2Config{
		// Endpoint on the authorization server to retrieve the access token from.
		TokenEndpoint: tokenEndpoint,
		// The endpoint redirect to for authorization in response to unauthorized requests.
		AuthorizationEndpoint: authorizationEndpoint,
		// Credentials used for OAuth.
		Credentials: credentials,
		// The redirect URI passed to the authorization endpoint. Supports header formatting
		// tokens. For more information, including details on header value syntax, see the
		// documentation on :ref:`custom request headers <config_http_conn_man_headers_custom_request_headers>`.
		//
		// This URI should not contain any query parameters.
		RedirectUri: redirectUri,
		// Matching criteria used to determine whether a path appears to be the result of a redirect from the authorization server.
		RedirectPathMatcher: redirectPathMatcher,
		// The path to sign a user out, clearing their credential cookies.
		SignoutPath: signoutPath,
		// Forward the OAuth token as a Bearer to upstream web service.
		ForwardBearerToken: forwardBearerToken,
		// // Any request that matches any of the provided matchers will be passed through without OAuth validation.
		// PassThroughMatcher: nil,
		// Optional list of OAuth scopes to be claimed in the authorization request. If not specified,
		// defaults to "user" scope.
		// OAuth RFC https://tools.ietf.org/html/rfc6749#section-3.3
		AuthScopes: authScopes,
		// Optional resource parameter for authorization request
		// RFC: https://tools.ietf.org/html/rfc8707
		Resources: resources,
	}
	oAuth2 := &envoy_extensions_filter_http_oauth2_v3.OAuth2{
		// Leave this empty to disable OAuth2 for a specific route, using per filter config.
		Config: config,
	}

	anyOAuth2, err := anypb.New(oAuth2)
	if err != nil {
		return nil, fmt.Errorf("auth.NewFilterHTTPOAuth2: cannot marshal filter oAuth2=%+#v, %w", oAuth2, err)
	}

	return anyOAuth2, nil
}

func GenerateHMAC() (string, error) {
	// Since HMAC use symmetric key algorithm, we can just generate random bytes as secret key.

	// Securely generate an HMAC of at least 32 bytes:
	// "$ head -c 32 /dev/urandom | base64"
	// As of yet, the Envoy's implementation uses a 32 bytes digest
	// (SHA-256) which makes 32 bytes for the secret a good choice.

	// Example HMAC
	// ```sh
	// $ head -c 32 /dev/urandom | base64
	// 7njsC6u31gWhLlGemUEr4YoPxa2i832PMPvlwABmD8Y=
	// ```

	// Random 32 bytes long string
	src := make([]byte, 32)
	_, err := rand.Read(src)
	if err != nil {
		return "", fmt.Errorf("auth.GenerateHMAC: failed on rand.Read(src), err=%w", err)
	}

	// Encode as base64 string
	hmac := base64.StdEncoding.EncodeToString(src)

	return hmac, nil
}
