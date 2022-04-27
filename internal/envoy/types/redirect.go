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
package types

import (
	"fmt"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

var redirectResponseCodeName = map[uint32]string{
	301: "MOVED_PERMANENTLY",
	302: "FOUND",
	303: "SEE_OTHER",
	307: "TEMPORARY_REDIRECT",
	308: "PERMANENT_REDIRECT",
}

type RouteRedirectBuilder struct {
	redirect *route.Route_Redirect
}

func NewRouteRedirectBuilder() *RouteRedirectBuilder {
	return &RouteRedirectBuilder{
		redirect: &route.Route_Redirect{
			Redirect: &route.RedirectAction{},
		},
	}
}

func (r *RouteRedirectBuilder) HostRedirect(host string) *RouteRedirectBuilder {
	r.redirect.Redirect.HostRedirect = host
	return r
}

func (r *RouteRedirectBuilder) PortRedirect(port uint32) *RouteRedirectBuilder {
	r.redirect.Redirect.PortRedirect = port
	return r
}

func (r *RouteRedirectBuilder) SchemeRedirect(scheme string) *RouteRedirectBuilder {
	if scheme != "" {
		r.redirect.Redirect.SchemeRewriteSpecifier =
			&route.RedirectAction_SchemeRedirect{SchemeRedirect: scheme}
	}

	return r
}

func (r *RouteRedirectBuilder) RegexRedirect(pattern, substitution string) *RouteRedirectBuilder {
	if pattern != "" {
		if r.redirect.Redirect == nil {
			r.redirect.Redirect = &route.RedirectAction{}
		}
		// If PathRedirect is set, this will overwrite it as they are mutually exclusive
		r.redirect.Redirect.PathRewriteSpecifier = &route.RedirectAction_RegexRewrite{
			RegexRewrite: GenerateRewriteRegex(pattern, substitution),
		}
	}

	return r
}

func GenerateRewriteRegex(pattern string, substitution string) *envoytypematcher.RegexMatchAndSubstitute {
	if pattern == "" {
		return nil
	}
	return &envoytypematcher.RegexMatchAndSubstitute{
		Pattern: &envoytypematcher.RegexMatcher{
			EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
				GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
			},
			Regex: pattern,
		},
		Substitution: substitution,
	}
}

func (r *RouteRedirectBuilder) PathRedirect(path string) *RouteRedirectBuilder {
	if path != "" {
		// If RegexRedirect is set, this will overwrite it as they are mutually exclusive
		r.redirect.Redirect.PathRewriteSpecifier = &route.RedirectAction_PathRedirect{
			PathRedirect: path,
		}
	}

	return r
}

func (r *RouteRedirectBuilder) ResponseCode(code uint32) *RouteRedirectBuilder {
	if code > 0 {
		if r.redirect.Redirect == nil {
			r.redirect.Redirect = &route.RedirectAction{}
		}
		// if the code is undefined, set it to 301 as default in Envoy
		// otherwise we need to convert HTTP code (e.g. 308) to internal go-control-plane enumeration
		if redirectActionResponseCodeName, ok := redirectResponseCodeName[code]; ok {
			r.redirect.Redirect.ResponseCode = route.RedirectAction_RedirectResponseCode(
				route.RedirectAction_RedirectResponseCode_value[redirectActionResponseCodeName],
			)
		} else {
			r.redirect.Redirect.ResponseCode = 301
		}
	}

	return r
}

func (r *RouteRedirectBuilder) StripQuery(stripQuery *bool) *RouteRedirectBuilder {
	if stripQuery != nil {
		r.redirect.Redirect.StripQuery = *stripQuery
	}

	return r
}

func (r *RouteRedirectBuilder) ValidateAndReturn() (*route.Route_Redirect, error) {
	if err := r.redirect.Redirect.Validate(); err != nil {
		return nil, fmt.Errorf("failed to build Route_Redirect: %w", err)
	}

	return r.redirect, nil
}
