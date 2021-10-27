package options

import (
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	HTTP_REDIRECT_MOVED_PERMANENTLY  uint32 = 301
	HTTP_REDIRECT_FOUND              uint32 = 302
	HTTP_REDIRECT_SEE_OTHER          uint32 = 303
	HTTP_REDIRECT_TEMPORARY_REDIRECT uint32 = 307
	HTTP_REDIRECT_PERMANENT_REDIRECT uint32 = 308
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

type RewriteRegex struct {
	Pattern      string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Substitution string `yaml:"substitution,omitempty" json:"substitution,omitempty"`
}

func (o *RedirectOptions) Validate() error {
	// TODO: add more validations
	// FIXME: currently this isn't called for suboptions!!!
	if o.PathRedirect != "" && o.RewriteRegex.Pattern != "" {
		return fmt.Errorf("path redirect and rewrite regex are defined but are mutually exclusive")
	}
	return v.ValidateStruct(o,
		v.Field(&o.SchemeRedirect, v.In("http", "https")),
		v.Field(&o.ResponseCode, v.In(
			HTTP_REDIRECT_FOUND,
			HTTP_REDIRECT_MOVED_PERMANENTLY,
			HTTP_REDIRECT_PERMANENT_REDIRECT,
			HTTP_REDIRECT_SEE_OTHER,
			HTTP_REDIRECT_TEMPORARY_REDIRECT,
		)),
		v.Field(&o.HostRedirect, is.Host),
		v.Field(&o.PortRedirect, is.Port),
	)
}
