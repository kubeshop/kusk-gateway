package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type QoSOptions struct {
	// Retries define how many times to retry calling the backend
	Retries uint32 `yaml:"retries,omitempty" json:"retries,omitempty"`
	// RequestTimeout is total request timeout
	RequestTimeout uint32 `yaml:"request_timeout,omitempty" json:"request_timeout,omitempty"`
	// IdleTimeout is timeout for idle connection
	IdleTimeout uint32 `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
}

func (o QoSOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Retries, v.Min(uint32(0)), v.Max(MaxRetries)),
		v.Field(&o.RequestTimeout, v.Min(uint32(0))),
		v.Field(&o.IdleTimeout, v.Min(uint32(0))),
	)
}
