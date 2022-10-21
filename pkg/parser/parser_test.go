package parser

import (
	"testing"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/kubeshop/kusk-gateway/pkg/options"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestMapRateLimitConf(t *testing.T) {
	rlOpt := &options.RateLimitOptions{
		RequestsPerUnit: 2,
		Unit:            "minute",
		PerConnection:   false,
		ResponseCode:    403,
	}
	out := mapRateLimitConf(rlOpt, "stat_prefix")

	want := &ratelimit.LocalRateLimit{
		StatPrefix: "stat_prefix",
		Status: &envoy_type_v3.HttpStatus{
			Code: envoy_type_v3.StatusCode(rlOpt.ResponseCode),
		},
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens: 2,
			TokensPerFill: &wrapperspb.UInt32Value{
				Value: 2,
			},
			FillInterval: &durationpb.Duration{
				Seconds: 60,
			},
		},
		FilterEnabled: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enabled",
		},
		FilterEnforced: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enforced",
		},
		Stage:                                 0,
		LocalRateLimitPerDownstreamConnection: rlOpt.PerConnection,
	}

	assert.Equal(t, want, out)
}

func TestMapRateLimitConfDefault(t *testing.T) {
	rlOpt := &options.RateLimitOptions{
		RequestsPerUnit: 2,
		Unit:            "minute",
	}
	out := mapRateLimitConf(rlOpt, "stat_prefix")

	want := &ratelimit.LocalRateLimit{
		StatPrefix: "stat_prefix",
		Status: &envoy_type_v3.HttpStatus{
			Code: envoy_type_v3.StatusCode(429),
		},
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens: 2,
			TokensPerFill: &wrapperspb.UInt32Value{
				Value: 2,
			},
			FillInterval: &durationpb.Duration{
				Seconds: 60,
			},
		},
		FilterEnabled: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enabled",
		},
		FilterEnforced: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enforced",
		},
		Stage:                                 0,
		LocalRateLimitPerDownstreamConnection: false,
	}

	assert.Equal(t, want, out)
}
