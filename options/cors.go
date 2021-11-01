package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type CORSOptions struct {
	Origins       []string `yaml:"origins,omitempty" json:"origins,omitempty"`
	Methods       []string `yaml:"methods,omitempty" json:"methods,omitempty"`
	Headers       []string `yaml:"headers,omitempty" json:"headers,omitempty"`
	ExposeHeaders []string `yaml:"expose_headers,omitempty" json:"expose_headers,omitempty"`

	// Pointer because default value of bool is false which could have unintended side effects
	// Check if not nil to ensure it's been set by user
	Credentials *bool `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	MaxAge      int   `yaml:"max_age,omitempty" json:"max_age,omitempty"`
}

func (o CORSOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Methods, v.Each(v.In("GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"))),
		v.Field(&o.MaxAge, v.Min(0)),
	)
}
