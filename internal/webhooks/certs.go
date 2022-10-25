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
package webhooks

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// CreateCertificates creates the self-signed CA and server certs for the Admission Webhook server.
// The expiration time for the certificates is 2 years for the really unusual cases when manager is not restarted in 2 years.
// Returns CA certificate to further patch Mutating and Validating configs.
func CreateCertificates(dnsNames []string, certsDirectory string, certFileName string, certKeyFileName string) ([]byte, error) {
	var caPEM, serverCertPEM, serverPrivKeyPEM *bytes.Buffer

	const (
		certOrganization string = "kusk.io"
	)
	// CA config.
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{certOrganization},
		},
		NotBefore: time.Now(),
		// 2 years expiration
		NotAfter:              time.Now().AddDate(2, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	// PEM encode CA cert
	caPEM = new(bytes.Buffer)
	if err := pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return nil, err
	}

	// server serverCert config
	serverCert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   dnsNames[0],
			Organization: []string{certOrganization},
		},
		NotBefore: time.Now(),
		// 2 years expiration
		NotAfter:     time.Now().AddDate(2, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, serverCert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	// PEM encode the  server cert and key
	serverCertPEM = new(bytes.Buffer)
	if err := pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	}); err != nil {
		return nil, err
	}

	serverPrivKeyPEM = new(bytes.Buffer)
	if err := pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	}); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(certsDirectory, 0755); err != nil {
		return nil, err
	}
	if err := writeFile(path.Join(certsDirectory, certFileName), serverCertPEM); err != nil {
		return nil, err
	}
	if err := writeFile(path.Join(certsDirectory, certKeyFileName), serverPrivKeyPEM); err != nil {
		return nil, err
	}

	return caPEM.Bytes(), nil
}

// writeFile writes data in the file at the given path
func writeFile(filepath string, sCert *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(sCert.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// GetWebhookServiceDNSNames locates the manager's webhook service with the predefined labels and
// returns the list of DNS names for this service to use during the certificates creation.
func GetWebhookServiceDNSNames(ctx context.Context, clientSet *kubernetes.Clientset) ([]string, error) {
	var dnsNames []string
	// Webhook service labels
	serviceLabels := map[string]string{
		"app.kubernetes.io/name":      "kusk-gateway",
		"app.kubernetes.io/component": "kusk-gateway-webhooks-service",
	}
	labelSelector, err := labels.ValidatedSelectorFromSet(serviceLabels)
	if err != nil {
		return dnsNames, fmt.Errorf("failed to form labels selector from map %v: %w", serviceLabels, err)
	}

	servicesList, err := clientSet.CoreV1().Services("").List(ctx, metav1.ListOptions{LabelSelector: labelSelector.String()})
	if err != nil {
		return dnsNames, fmt.Errorf("failed getting services from the cluster with labels %v: %w", serviceLabels, err)
	}

	switch svcs := len(servicesList.Items); {
	case svcs == 0:
		return dnsNames, fmt.Errorf("no service detected in the cluster when searching with the labels %s", serviceLabels)
	case svcs > 1:
		return dnsNames, fmt.Errorf("more than one service detected in the cluster when searching with the labels %s", serviceLabels)
	}
	service := servicesList.Items[0]
	serviceName := service.GetName()
	serviceNamespace := service.GetNamespace()

	dnsNames = append(dnsNames, serviceName)
	dnsNames = append(dnsNames, strings.Join([]string{serviceName, serviceNamespace}, "."))
	dnsNames = append(dnsNames, strings.Join([]string{serviceName, serviceNamespace, "svc"}, "."))
	dnsNames = append(dnsNames, strings.Join([]string{serviceName, serviceNamespace, "svc", "cluster", "local"}, "."))

	return dnsNames, nil
}

// The necessary RBAC permissions for the manager to patch mutating webhooks
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;create;update;patch
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;create;update;patch

// UpdateWebhookConfiguration patches the Kusk Gateway Mutating and Validating Admission configurations with the certificate authority certificate.
func UpdateWebhookConfiguration(ctx context.Context, clientSet *kubernetes.Clientset, caPemCert []byte) error {
	// Labels to find the configurations
	configLabels := map[string]string{"app.kubernetes.io/name": "kusk-gateway"}
	labelSelector, err := labels.ValidatedSelectorFromSet(configLabels)
	if err != nil {
		return fmt.Errorf("failed to form labels selector from map %v: %w", configLabels, err)
	}

	// Mutating configs patching
	mutatingConfigList, err := clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().List(ctx, metav1.ListOptions{LabelSelector: labelSelector.String()})
	if err != nil {
		return fmt.Errorf("failed getting mutating webhook configurations from the cluster with labels %v: %w", configLabels, err)
	}
	for _, mutatingConfig := range mutatingConfigList.Items {
		patch := generateWebhooksJsonPatch(caPemCert, len(mutatingConfig.Webhooks))
		if _, err := clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Patch(ctx, mutatingConfig.Name, types.JSONPatchType, patch, metav1.PatchOptions{}); err != nil {
			return fmt.Errorf("failure patching mutating webhook configuration %s: %v", mutatingConfig.Name, err)
		}
	}

	// Validating configs patching
	validatingConfigList, err := clientSet.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{LabelSelector: labelSelector.String()})
	if err != nil {
		return fmt.Errorf("failed getting validating webhook configurations from the cluster with labels %v: %w", configLabels, err)
	}
	for _, validatingConfig := range validatingConfigList.Items {
		patch := generateWebhooksJsonPatch(caPemCert, len(validatingConfig.Webhooks))
		if _, err := clientSet.AdmissionregistrationV1().ValidatingWebhookConfigurations().Patch(ctx, validatingConfig.Name, types.JSONPatchType, patch, metav1.PatchOptions{}); err != nil {
			return fmt.Errorf("failure patching validating webhook configuration %s: %v", validatingConfig.Name, err)
		}
	}
	return nil
}

// generateWebhooksJsonPatch returns JSON patch to update CABundle in Admission configuration webhooks
func generateWebhooksJsonPatch(caBundle []byte, count int) []byte {
	operations := make([]string, count, count)
	b64caBundle := base64.StdEncoding.EncodeToString(caBundle)
	webhookpatch := func(number int) string {
		return fmt.Sprintf(`{"op": "replace", "path": "/webhooks/%d/clientConfig/caBundle", "value":"%s"}`, number, b64caBundle)
	}
	for i := 0; i < count; i++ {
		operations[i] = webhookpatch(i)
	}
	patchString := fmt.Sprintf("[%s]", strings.Join(operations, ","))
	return []byte(patchString)

}
