package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ServiceOptions struct {

	// Name is the upstream Service's name.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Port is the upstream Service's port.
	Port uint32 `yaml:"port,omitempty" json:"port,omitempty"`
}

func (o ServiceOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Name, is.DNSName),
		v.Field(&o.Port, v.Min(uint32(1)), v.Max(uint32(65356)), v.When(o.Name != "", v.Required)),
	)
}
