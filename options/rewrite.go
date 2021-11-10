package options

import (
	"fmt"
)

// RewriteRegex is used during the redirects and paths mangling
type RewriteRegex struct {
	Pattern      string `yaml:"pattern,omitempty" json:"pattern"`
	Substitution string `yaml:"substitution,omitempty" json:"substitution"`
}

func (r RewriteRegex) Validate() error {
	if r.Substitution != "" && r.Pattern == "" {
		return fmt.Errorf("rewrite regex pattern must be specified if the substitution is defined")
	}
	return nil
}
