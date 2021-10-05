package options

import (
	"github.com/go-ozzo/ozzo-validation/v4"
)

type PathOptions struct {
	// Base is the preceding prefix for the route (i.e. /your-prefix/here/rest/of/the/route).
	// Default value is "/".
	Base string `yaml:"base,omitempty" json:"base,omitempty"`

	// TrimPrefix is the prefix that would be omitted from the URL when request is being forwarded
	// to the upstream service, i.e. given that Base is set to "/petstore/api/v3", TrimPrefix is set to "/petstore",
	// path that would be generated is "/petstore/api/v3/pets", URL that the upstream service would receive
	// is "/api/v3/pets".
	TrimPrefix string `yaml:"trim_prefix,omitempty" json:"trim_prefix,omitempty"`

	// Split forces Kusk to generate a separate resource for each Path or Operation, where appropriate.
	Split bool `yaml:"split,omitempty" json:"split,omitempty"`
}

func (o *PathOptions) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Base, validation.Required.Error("Base path required")),
	)
}
