package options

type RateLimitOptions struct {
	RPS   uint32 `json:"rps,omitempty" yaml:"rps,omitempty"`
	Burst uint32 `json:"burst,omitempty" yaml:"burst,omitempty"`
	Group string `json:"group,omitempty" yaml:"group,omitempty"`
}

func (o RateLimitOptions) Validate() error {
	// TODO: write validations
	return nil
}
