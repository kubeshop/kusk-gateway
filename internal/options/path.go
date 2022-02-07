package options

import (
	"fmt"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

const MaxRetries uint32 = 255

type PathOptions struct {
	// Prefix is the preceding prefix for the route (i.e. /your-prefix/here/rest/of/the/route).
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`

	// Rewrite is the pattern (regex) and a substitution string that will change URL when request is being forwarded
	// to the upstream service.
	// e.g. given that Prefix is set to "/petstore/api/v3", and with
	// Rewrite.Pattern is set to "^/petstore", Rewrite.Substitution is set to ""
	// path that would be generated is "/petstore/api/v3/pets", URL that the upstream service would receive
	// is "/api/v3/pets".
	Rewrite RewriteRegex `yaml:"rewrite,omitempty" json:"rewrite,omitempty"`
}

func (o PathOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Prefix, v.By(validatePrefix)),
		v.Field(&o.Rewrite),
	)
}

func validatePrefix(value interface{}) error {
	path, ok := value.(string)
	if !ok {
		return fmt.Errorf("validatable object must be a string")
	}
	if path == "" {
		return nil
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("prefix path should start with /")
	}
	return nil
}
