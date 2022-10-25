/*
MIT License

# Copyright (c) 2022 Kubeshop

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
package cert

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

var (
	pemStart      = []byte("-----BEGIN ")
	certBlockType = "CERTIFICATE"
)

// DecodeCertificates accepts a byte slice of data and decodes it into a slice of Certificates
// if certificates can't be read or they are invalid, an error is returned
// Taken from https://github.com/zakjan/cert-chain-resolver/blob/master/certUtil/io.go
func DecodeCertificates(data []byte) ([]*x509.Certificate, error) {
	if isPEM := bytes.HasPrefix(data, pemStart); isPEM {
		var certs []*x509.Certificate

		for len(data) > 0 {
			var block *pem.Block

			block, data = pem.Decode(data)
			if block == nil {
				return nil, fmt.Errorf("a valid block wasn't found in byte data")
			}
			if block.Type != certBlockType {
				return nil, fmt.Errorf("invalid block type for detected PEM cert")
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, errors.New("invalid certificate")
			}

			certs = append(certs, cert)
		}

		return certs, nil
	}

	certs, err := x509.ParseCertificates(data)
	if err != nil {
		return nil, errors.New("invalid certificate")
	}

	return certs, nil
}
