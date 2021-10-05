package options

import "reflect"

type TimeoutOptions struct {
	// RequestTimeout is total request timeout
	RequestTimeout uint32 `yaml:"request_timeout,omitempty" json:"request_timeout,omitempty"`
	// IdleTimeout is timeout for idle connection
	IdleTimeout uint32 `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
}

func (o *Options) GetTimeoutOpts(path, method string) TimeoutOptions {
	// take global timeout options
	timeoutOpts := o.Timeouts

	// if non-zero path-level timeout options are different, override them
	if pathSubOpts, ok := o.PathSubOptions[path]; ok {
		if !reflect.DeepEqual(TimeoutOptions{}, pathSubOpts.Timeouts) &&
			!reflect.DeepEqual(timeoutOpts, pathSubOpts.Timeouts) {
			timeoutOpts = pathSubOpts.Timeouts
		}
	}

	// if non-zero operation-level timeout options are different, override them
	if opSubOpts, ok := o.OperationSubOptions[path]; ok {
		if !reflect.DeepEqual(TimeoutOptions{}, opSubOpts.Timeouts) &&
			!reflect.DeepEqual(timeoutOpts, opSubOpts.Timeouts) {
			timeoutOpts = opSubOpts.Timeouts
		}
	}

	return timeoutOpts
}

func (o *TimeoutOptions) Validate() error {
	return nil
}
