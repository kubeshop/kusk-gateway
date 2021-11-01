package options

import (
	"fmt"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
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

func (o PathOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Base, v.By(checkBasePath)),
		v.Field(&o.Retries, v.Min(uint32(0)), v.Max(uint32(255))),
		v.Field(&o.Rewrite),
	)
}

func checkBasePath(value interface{}) error {
	path, ok := value.(string)
	if !ok {
		return fmt.Errorf("validatable object must be the string")
	}
	if path == "" {
		return nil
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("base path should start with /")
	}
	return nil
}
