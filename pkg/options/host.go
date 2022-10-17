/*
MIT License

Copyright (c) 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package options

import (
	"fmt"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Host is a vHost (and domain name) definition that is used during request routing.
// Could be wildcard ("*" - all vhosts), prefix/suffix wildcard (*-example.com, example.*, but not both *example*),
// or simple domain (www.example.com)
type Host string

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
	// Finally check if this is a valid Host
	return v.Validate(host, is.Host)
}
