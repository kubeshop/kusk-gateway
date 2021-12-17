package types

import (
	"reflect"
	"testing"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func TestGetHeaderMatcherConfig(t *testing.T) {
	type args struct {
		methods []string
		cors    bool
	}
	tests := []struct {
		name string
		args args
		want *route.HeaderMatcher
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHeaderMatcherConfig(tt.args.methods, tt.args.cors); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHeaderMatcherConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
