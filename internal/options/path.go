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
}

func (o PathOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Prefix, v.By(validatePrefix)),
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
