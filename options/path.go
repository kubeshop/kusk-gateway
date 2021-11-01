package options

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type PathOptions struct {
	// Base is the preceding prefix for the route (i.e. /your-prefix/here/rest/of/the/route).
	Base string `yaml:"base,omitempty" json:"base,omitempty"`

	// Rewrite is the pattern (regex) and a substitution string that will change URL when request is being forwarded
	// to the upstream service.
	// e.g. given that Base is set to "/petstore/api/v3", and with
	// Rewrite.Pattern is set to "^/petstore", Rewrite.Substitution is set to ""
	// path that would be generated is "/petstore/api/v3/pets", URL that the upstream service would receive
	// is "/api/v3/pets".
	Rewrite RewriteRegex `yaml:"rewrite,omitempty" json:"rewrite,omitempty"`
	Retries uint32       `yaml:"retries,omitempty" json:"retries,omitempty"`
}

func (o *PathOptions) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Base, validation.Required.Error("Base path required")),
	)
}
