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

type OAuthOptions struct {
	AuthScopes            []string    `json:"auth_scopes,omitempty" yaml:"auth_scopes,omitempty"`
	Credentials           Credentials `json:"unit,omitempty" yaml:"unit,omitempty"`
	RedirectURI           string      `json:"redirect_uri,omitempty" yaml:"redirect_uri,omitempty"`
	AuthorizationEndpoint string      `json:"authorization_endpoint"`
	TokenEndpoint         struct {
		Cluster string `json:"cluster"`
		URI     string `json:"uri"`
		Timeout string `json:"timeout"`
	} `json:"token_endpoint"`

	RedirectPathMatcher struct {
		Path Path `json:"path"`
	} `yaml:"redirect_path_matcher,omitempty" json:"redirect_path_matcher,omitempty"`
	SignoutPath struct {
		Path Path `json:"path"`
	} `json:"signout_path"`
}

type Credentials struct {
	ClientID    string `json:"client_id"`
	TokenSecret Secret `json:"token_secret"`
	HMACSecret  Secret `json:"hmac_secret`
}

type Path struct {
	Exact string `json:"exact"`
}
type Secret struct {
	Name      string `json:"name"`
	SDSConfig struct {
		Path string `json:"path"`
	} `json:"sds_config"`
}

func (o OAuthOptions) Validate() error {
	return nil
}
