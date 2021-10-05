package options

import "reflect"

type RateLimitOptions struct {
	RPS   uint32 `json:"rps,omitempty" yaml:"rps,omitempty"`
	Burst uint32 `json:"burst,omitempty" yaml:"burst,omitempty"`
	Group string `json:"group,omitempty" yaml:"group,omitempty"`
}

func (o *Options) GetRateLimitOpts(path, method string) RateLimitOptions {
	// take global rate limit options
	rateLimitOpts := o.RateLimits

	// if non-zero path-level rate limit options are different, override with them
	if pathSubOpts, ok := o.PathSubOptions[path]; ok &&
		pathSubOpts.RateLimits.ShouldOverride(rateLimitOpts) {
		rateLimitOpts = pathSubOpts.RateLimits
	}

	// if non-zero operation-level rate limit options are different, override them
	if opSubOpts, ok := o.OperationSubOptions[path]; ok &&
		opSubOpts.RateLimits.ShouldOverride(rateLimitOpts) {
		rateLimitOpts = opSubOpts.RateLimits
	}

	return rateLimitOpts
}

func (o *RateLimitOptions) ShouldOverride(opts RateLimitOptions) bool {
	return !reflect.DeepEqual(CORSOptions{}, o) && !reflect.DeepEqual(opts, o)
}

func (o *RateLimitOptions) Validate() error {
	return nil
}
