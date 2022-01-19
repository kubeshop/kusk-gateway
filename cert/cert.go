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
