package options

type MockingOptions struct {
	Enabled *bool `json:"enabled" yaml:"mocking"`
}

func (o MockingOptions) Validate() error {
	// Nothing to validate yet
	return nil
}
