package options

type ValidationOptions struct {
	Request *RequestValidationOptions `json:"request,omitempty" yaml:"request,omitempty"`
}

type RequestValidationOptions struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}
