package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kubeshop/kusk-gateway/options"
)

func generateCORSPolicy(corsOpts *options.CORSOptions) (*route.CorsPolicy, error) {
	if reflect.DeepEqual(&options.CORSOptions{}, corsOpts) {
		return nil, nil
	}
	allowOriginsMatcher := []*envoytypematcher.StringMatcher{}
	for _, origin := range corsOpts.Origins {
		entry := &envoytypematcher.StringMatcher{
			// TODO: We support only exact strings, no regexp - fix this if applicable
			MatchPattern: &envoytypematcher.StringMatcher_Exact{
				Exact: origin,
			},
			IgnoreCase: false,
		}
		allowOriginsMatcher = append(allowOriginsMatcher, entry)
	}
	corsPolicy := &route.CorsPolicy{
		AllowOriginStringMatch: allowOriginsMatcher,
		AllowMethods:           strings.Join(corsOpts.Methods, ","),
		AllowHeaders:           strings.Join(corsOpts.Headers, ","),
		ExposeHeaders:          strings.Join(corsOpts.ExposeHeaders, ","),
		MaxAge:                 strconv.Itoa(corsOpts.MaxAge),
	}
	if corsOpts.Credentials != nil {
		corsPolicy.AllowCredentials = &wrapperspb.BoolValue{
			Value: *corsOpts.Credentials,
		}
	}
	if err := corsPolicy.Validate(); err != nil {
		return nil, fmt.Errorf("incorrect CORS configuration: %w", err)
	}
	return corsPolicy, nil
}
