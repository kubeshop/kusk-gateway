package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type CORSOptions struct {
	Origins       []string `yaml:"origins,omitempty" json:"origins,omitempty"`
	Methods       []string `yaml:"methods,omitempty" json:"methods,omitempty"`
	Headers       []string `yaml:"headers,omitempty" json:"headers,omitempty"`
	ExposeHeaders []string `yaml:"expose_headers,omitempty" json:"expose_headers,omitempty"`

	// Bool is a pointer because default value of bool is false which could have unintended side effects
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

// DeepCopy creates a copy of an object
func (in *CORSOptions) DeepCopy() *CORSOptions {
	if in == nil {
		return nil
	}
	out := new(CORSOptions)
	*out = *in
	if in.Origins != nil {
		in, out := &in.Origins, &out.Origins
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Methods != nil {
		in, out := &in.Methods, &out.Methods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExposeHeaders != nil {
		in, out := &in.ExposeHeaders, &out.ExposeHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return out
}
