package cert

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

func FromBytes(data []byte) (*x509.Certificate, error) {
	cert, err := decodeCertificatePEM(data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode certificate: %w", err)
	}

	return cert, nil
}

func decodeCertificatePEM(data []byte) (*x509.Certificate, error) {
	const certificateBlockType = "CERTIFICATE"

	var block *pem.Block
	for {
		block, data = pem.Decode(data)
		if block == nil {
			return nil, errors.New("failed to parse certificate PEM")
		}
		// append only certificates
		if block.Type == certificateBlockType {
			return x509.ParseCertificate(block.Bytes)
		}
		if len(data) == 0 {
			break
		}
	}
	return nil, fmt.Errorf("data did not contain a valid certificate")
}
