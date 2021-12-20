package types

import (
	"reflect"
	"testing"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

func TestNewVirtualHost(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *VirtualHost
	}{
		{
			name: "get new virtual host",
			args: args{
				name: "*",
			},
			want: &VirtualHost{
				&route.VirtualHost{
					Name:    "*",
					Domains: []string{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewVirtualHost(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewVirtualHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVirtualHost_AddRoute(t *testing.T) {
	type fields struct {
		VirtualHost *route.VirtualHost
	}
	type args struct {
		r *route.Route
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "add route",
			fields: fields{
				VirtualHost: &route.VirtualHost{
					Name:    "virtualhost",
					Domains: []string{"*"},
				},
			},
			args: args{
				r: &route.Route{
					Name: "/path-GET",
				},
			},
		},
		{
			name: "add route that already exists",
			fields: fields{
				VirtualHost: &route.VirtualHost{
					Name:    "virtualhost",
					Domains: []string{"*"},
					Routes: []*route.Route{
						{
							Name: "/path-GET",
						},
					},
				},
			},
			args: args{
				r: &route.Route{
					Name: "/path-GET",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VirtualHost{
				VirtualHost: tt.fields.VirtualHost,
			}
			if err := v.AddRoute(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("AddRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
