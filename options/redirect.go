package options

import (
	"fmt"

	"net/http"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// RedirectOptions provide http redirects
// see https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-redirectaction
type RedirectOptions struct {
	// http|https
	SchemeRedirect string `yaml:"scheme_redirect,omitempty" json:"scheme_redirect,omitempty"`

	HostRedirect string `yaml:"host_redirect,omitempty" json:"host_redirect,omitempty"`
	PortRedirect uint32 `yaml:"port_redirect,omitempty" json:"port_redirect,omitempty"`

	// These 2 are mutually exclusive
	PathRedirect string       `yaml:"path_redirect,omitempty" json:"path_redirect,omitempty"`
	RewriteRegex RewriteRegex `yaml:"rewrite_regex,omitempty" json:"rewrite_regex,omitempty"`

	ResponseCode uint32 `yaml:"response_code,omitempty" json:"response_code,omitempty"`
	StripQuery   *bool  `yaml:"strip_query,omitempty" json:"strip_query,omitempty"`
}

func (o *RedirectOptions) Validate() error {
	// TODO: add more validations
	if o.PathRedirect != "" && o.RewriteRegex.Pattern != "" {
		return fmt.Errorf("path redirect and rewrite regex are defined but are mutually exclusive")
	}
	return v.ValidateStruct(o,
		v.Field(&o.SchemeRedirect, v.In("http", "https")),
		v.Field(&o.ResponseCode, v.In(
			http.StatusFound,
			http.StatusMovedPermanently,
			http.StatusPermanentRedirect,
			http.StatusSeeOther,
			http.StatusTemporaryRedirect,
		)),
		v.Field(&o.HostRedirect, is.Host),
		v.Field(&o.PortRedirect, is.Port),
	)
}
