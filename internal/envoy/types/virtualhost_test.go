/*
MIT License

# Copyright (c) 2022 Kubeshop

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
