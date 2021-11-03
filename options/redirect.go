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

	// mutually exclusive with rewrite regex redirect
	PathRedirect string `yaml:"path_redirect,omitempty" json:"path_redirect,omitempty"`
	// mutually exclusive with path redirect
	RewriteRegex *RewriteRegex `yaml:"rewrite_regex,omitempty" json:"rewrite_regex,omitempty"`

	ResponseCode uint32 `yaml:"response_code,omitempty" json:"response_code,omitempty"`
	StripQuery   *bool  `yaml:"strip_query,omitempty" json:"strip_query,omitempty"`
}

func (o RedirectOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.SchemeRedirect, v.In("http", "https")),
		v.Field(&o.ResponseCode, v.In(
			uint32(http.StatusFound),
			uint32(http.StatusMovedPermanently),
			uint32(http.StatusPermanentRedirect),
			uint32(http.StatusSeeOther),
			uint32(http.StatusTemporaryRedirect),
		)),
		v.Field(&o.HostRedirect, is.Host),
		v.Field(&o.PortRedirect, v.Min(uint32(1)), v.Max(uint32(65535))),
		v.Field(&o.RewriteRegex),
		v.Field(&o.PathRedirect, v.By(o.MutuallyExlusivePathRedirectCheck)),
	)
}

// MutuallyExclusivePathRedirectCheck returns error if both path redirect and regex substitution redirect are enabled
func (o RedirectOptions) MutuallyExlusivePathRedirectCheck(value interface{}) error {
	pathRedirect, ok := value.(string)
	if !ok {
		return fmt.Errorf("validatable object must be a string")
	}
	if pathRedirect != "" && o.RewriteRegex.Pattern != "" {
		return fmt.Errorf("only one of path or rewrite regex redirects may be specified, but supplied both")
	}
	return nil
}

// DeepCopy is used to copy
func (in *RedirectOptions) DeepCopy() *RedirectOptions {
	if in == nil {
		return nil
	}
	out := new(RedirectOptions)
	*out = *in
	if in.RewriteRegex != nil {
		*out = *in
	}
	return out
}
