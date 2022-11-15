package options

type DevPortalOptions struct {
	Path    string `json:"path,omitempty" yaml:"path,omitempty"`
	LogoUrl string `json:"logo_url,omitempty" yaml:"logo_url,omitempty"`
	Title   string `json:"title,omitempty" yaml:"title,omitempty"`
}

func (o *DevPortalOptions) Validate() error {
	return nil
}
