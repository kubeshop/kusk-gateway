package config

import (
	"fmt"
	"reflect"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"github.com/kubeshop/kusk-gateway/options"
)

var redirectResponseCodeName = map[uint32]string{
	301: "MOVED_PERMANENTLY",
	302: "FOUND",
	303: "SEE_OTHER",
	307: "TEMPORARY_REDIRECT",
	308: "PERMANENT_REDIRECT",
}

func generateRedirectAction(redirectOpts *options.RedirectOptions) (*route.RedirectAction, error) {
	if reflect.DeepEqual(&options.RedirectOptions{}, redirectOpts) {
		return nil, nil
	}
	redirectAction := &route.RedirectAction{
		HostRedirect: redirectOpts.HostRedirect,
		PortRedirect: redirectOpts.PortRedirect,
	}
	if redirectOpts.SchemeRedirect != "" {
		redirectAction.SchemeRewriteSpecifier = &route.RedirectAction_SchemeRedirect{SchemeRedirect: redirectOpts.SchemeRedirect}
	}
	// PathRedirect and RewriteRegex are mutually exlusive
	// Path rewrite with regex
	if redirectOpts.RewriteRegex.Pattern != "" {
		redirectAction.PathRewriteSpecifier = &route.RedirectAction_RegexRewrite{
			RegexRewrite: generateRewriteRegex(redirectOpts.RewriteRegex.Pattern, redirectOpts.RewriteRegex.Substitution),
		}
	}
	// Or path rewrite with path redirect
	if redirectOpts.PathRedirect != "" {
		redirectAction.PathRewriteSpecifier = &route.RedirectAction_PathRedirect{
			PathRedirect: redirectOpts.PathRedirect,
		}
	}
	// if the code is undefined, it is set to 301 by default in Envoy
	// otherwise we need to convert HTTP code (e.g. 308) to internal go-control-plane enumeration
	if redirectOpts.ResponseCode != 0 {
		// go-control-plane uses internal map of response code names that we need to translate from real HTTP code
		redirectActionResponseCodeName, ok := redirectResponseCodeName[redirectOpts.ResponseCode]
		if !ok {
			return nil, fmt.Errorf("missing redirect code name for HTTP code %d", redirectOpts.ResponseCode)
		}
		code := route.RedirectAction_RedirectResponseCode_value[redirectActionResponseCodeName]
		redirectAction.ResponseCode = route.RedirectAction_RedirectResponseCode(code)
	}
	if redirectOpts.StripQuery != nil {
		redirectAction.StripQuery = *redirectOpts.StripQuery
	}
	if err := redirectAction.Validate(); err != nil {
		return nil, fmt.Errorf("incorrect Redirect Action: %w", err)
	}
	return redirectAction, nil
}
