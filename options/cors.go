package options

import "reflect"

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

func (o *Options) GetCORSOpts(path, method string) CORSOptions {
	// take global CORS options
	corsOpts := o.CORS

	// if non-zero path-level CORS options are different, override with them
	if pathSubOpts, ok := o.PathSubOptions[path]; ok {
		if !reflect.DeepEqual(CORSOptions{}, pathSubOpts.CORS) &&
			!reflect.DeepEqual(corsOpts, pathSubOpts.CORS) {
			corsOpts = pathSubOpts.CORS
		}
	}

	// if non-zero operation-level CORS options are different, override them
	if opSubOpts, ok := o.OperationSubOptions[path]; ok {
		if !reflect.DeepEqual(CORSOptions{}, opSubOpts.CORS) &&
			!reflect.DeepEqual(corsOpts, opSubOpts.CORS) {
			corsOpts = opSubOpts.CORS
		}
	}

	return corsOpts
}

func (o *CORSOptions) Validate() error {
	return nil
}
