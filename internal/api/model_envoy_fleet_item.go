/*
 * Kusk Gateway API
 *
 * This is the Kusk Gateway Management API
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api

type EnvoyFleetItem struct {
	Name string `json:"name"`

	Namespace string `json:"namespace"`

	Apis []ApiItemFleet `json:"apis,omitempty"`

	Services []ServiceItem `json:"services,omitempty"`

	StaticRoutes []StaticRouteItemFleet `json:"staticRoutes,omitempty"`
}

// AssertEnvoyFleetItemRequired checks if the required fields are not zero-ed
func AssertEnvoyFleetItemRequired(obj EnvoyFleetItem) error {
	elements := map[string]interface{}{
		"name":      obj.Name,
		"namespace": obj.Namespace,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.Apis {
		if err := AssertApiItemFleetRequired(el); err != nil {
			return err
		}
	}
	for _, el := range obj.Services {
		if err := AssertServiceItemRequired(el); err != nil {
			return err
		}
	}
	for _, el := range obj.StaticRoutes {
		if err := AssertStaticRouteItemFleetRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseEnvoyFleetItemRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of EnvoyFleetItem (e.g. [][]EnvoyFleetItem), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseEnvoyFleetItemRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aEnvoyFleetItem, ok := obj.(EnvoyFleetItem)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertEnvoyFleetItemRequired(aEnvoyFleetItem)
	})
}
