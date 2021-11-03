package options

import (
	"fmt"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Host is a vhost (domain name) definition that is used during request routing.
// Could be wildcard ("*" - all vhosts), prefix/suffix wildcard (*-example.com, example.*, but not both *example*),
// or simple domain (www.example.com)
type Host string

// validation helper to check vHost
func (h Host) Validate() error {
	host := string(h)
	// We're looking for *, either single, in the preffix and in the suffix.
	// Similarly to Envoy we support preffix and suffix wildcard (*), single wildcard and usual hostnames
	// For prefix and suffix * is replaced with "w" to form a DNS name that we can validate (e.g. "*-site.com", "site.*").
	switch {
	case host == "*":
		return nil
	case strings.HasPrefix(host, "*") && strings.HasSuffix(host, "*"):
		return fmt.Errorf("wildcards are not supported on both the start AND the end of hostname")
	case strings.HasPrefix(host, "*"):
		host = strings.TrimPrefix(host, "*")
		host = "w" + host
	case strings.HasSuffix(host, "*"):
		host = strings.TrimSuffix(host, "*")
		host = host + "w"
	}
	// Finally check if this is a valid DNS name
	return v.Validate(host, is.DNSName)
}
