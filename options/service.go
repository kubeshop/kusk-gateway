package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type ServiceOptions struct {
	// Namespace is the namespace containing the upstream Service.
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`

	// Name is the upstream Service's name.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Port is the upstream Service's port. Default value is 80.
	Port int32 `yaml:"port,omitempty" json:"port,omitempty"`
}

func (o *ServiceOptions) Validate() error {
	return v.ValidateStruct(o,
		v.Field(&o.Namespace, v.Required.Error("service.namespace is required")),
		v.Field(&o.Name, v.Required.Error("service.name is required")),
		v.Field(&o.Port, v.Required.Error("service.port is required"), v.Min(1), v.Max(65535)),
	)
}
