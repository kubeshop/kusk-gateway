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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeCertificates(t *testing.T) {
	singleCert := `-----BEGIN CERTIFICATE-----
MIIC1TCCAb2gAwIBAgIRAIJmUfoLsdqXVQqzT1CTMNUwDQYJKoZIhvcNAQELBQAw
ADAeFw0yMjAxMTcxNDIzNDVaFw0yMjA0MTcxNDIzNDVaMAAwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQC/MvDaXxSxdO3K3L5PY/OP9Ol7juwnPtOi651R
J7S3r2FTmZB6zUMRJG0oGFjfCQheXZJQkxURmSfdW/tkzRWl4Bme8xh4kFNdi/3t
ddCE2ckNvp9UCaxT8baRiG+xT/7TAONK8XoDLIyH2/zpprtVE0xo38VmWYmmgpNM
VEf87SXCSkO/fGW6Pt1qUwu47I4/5jQRh9B+SJQwmmyvR55RQ1Z9otCwzNgOteV0
0Jn39fgCkavEIwsUwyV6hE2zjpl0uTkw93cHbn2mJY6sAElLeRZYm2Xo/2Jt0BOZ
+3pfV/yHaXLg+/eZYHE7wcYcLGCsjFbM43PLAhr8mUR93Y0FAgMBAAGjSjBIMA4G
A1UdDwEB/wQEAwIFoDAMBgNVHRMBAf8EAjAAMCgGA1UdEQEB/wQeMByCC3RvZG9t
dmMuY29tgg1teXRvZG9tdmMuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQCZuXvIrx/a
pdvC2ACppazqtvE+WA4EZZlxFgk3zCgkhNBFIFfAJq5F5uGLAzhgrnxvcYk2kfqx
Ne/uCskl5en2gcd0zNyyJxPLUI4nCSlNje8RK9k80mlYh5GOeFUSmKgx2afn0dYI
aLWEgNOHbxJM+mEBGyLL0z9ps5ypxin6BfjyDy6rfXZHINGXbIpfaURuYhawMteW
MsetexKIFgJdt0J62XJvPuQpj58mSLZaDLf1lAtdVssg6Kl3Ev3EXzEaOYm2Xgef
hxKR99RwftrUXUWusQa/jjUB2JQYh0g3c9L4FoCRiLt2mYL/8JM8ihNqGheu+IGx
0Z7hvxeupgPG
-----END CERTIFICATE-----`

	certs, err := DecodeCertificates([]byte(singleCert))
	assert.Nil(t, err)
	assert.Len(t, certs, 1, "expected single certificate")
	assert.Len(t, certs[0].DNSNames, 2, "expected 2 DNS names")
}
