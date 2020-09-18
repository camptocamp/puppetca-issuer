/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provisioners

import (
	"context"
	"crypto/x509"
	"fmt"
	"sync"

	"github.com/camptocamp/go-puppetca/puppetca"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"k8s.io/apimachinery/pkg/types"
)

var collection = new(sync.Map)

// PuppetCA implements a Puppet CA provisioner in charge of signing certificate
// requests by calling Puppet CA API's
type PuppetCA struct {
	name        string
	provisioner *PuppetCAProvisioner
}

type PuppetCAProvisioner struct {
	client *puppetca.Client
	url    string
	cert   string
	key    string
	caCert string
}

func NewProvisioner(url string,
	cert string, key string, caCert string) (p *PuppetCAProvisioner) {

	return &PuppetCAProvisioner{
		url: url, cert: cert, key: key, caCert: caCert,
	}
}

// Load returns a Step provisioner by NamespacedName.
func Load(namespacedName types.NamespacedName) (*PuppetCAProvisioner, bool) {
	v, ok := collection.Load(namespacedName)
	if !ok {
		return nil, ok
	}
	p, ok := v.(*PuppetCAProvisioner)
	return p, ok
}

// Store adds a new provisioner to the collection by NamespacedName.
func Store(namespacedName types.NamespacedName, provisioner *PuppetCAProvisioner) {
	collection.Store(namespacedName, provisioner)
}

// Sign sends the certificate requests to the Step CA and returns the signed
// certificate.
func (p *PuppetCAProvisioner) Sign(ctx context.Context, cr *certmanager.CertificateRequest) ([]byte, []byte, error) {
	if p.client == nil {
		client, err := puppetca.NewClient(p.url, p.key, p.cert, p.caCert)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to initialize Puppet CA client")
		}
		p.client = &client
	}

	return nil, nil, fmt.Errorf("Failed to sign")
	/*
		// decode and check certificate request
		csr, err := decodeCSR(cr.Spec.CSRPEM)
		if err != nil {
			return nil, nil, err
		}

		sans := append([]string{}, csr.DNSNames...)
		for _, ip := range csr.IPAddresses {
			sans = append(sans, ip.String())
		}

		subject := csr.Subject.CommonName
		if subject == "" {
			subject = generateSubject(sans)
		}

		sess := session.Must(session.NewSession(&aws.Config{
			MaxRetries: aws.Int(3),
		}))

		svc := acmpca.New(sess, &aws.Config{
			Region: aws.String(p.region),
			Credentials: credentials.NewStaticCredentials(p.accesskey,
				p.secretkey, ""),
		})

		cparams := acmpca.IssueCertificateInput{
			CertificateAuthorityArn: aws.String(p.arn),
			SigningAlgorithm:        aws.String(acmpca.SigningAlgorithmSha256withrsa),
			Csr:                     cr.Spec.CSRPEM,
			Validity: &acmpca.Validity{
				Type:  aws.String(acmpca.ValidityPeriodTypeDays),
				Value: aws.Int64(int64(cr.Spec.Duration.Hours() / 24)),
			},
			IdempotencyToken: aws.String("awspca"),
		}

		output, err := svc.IssueCertificate(&cparams)

		if err != nil {
			return nil, nil, err
		}

		// wait for cert

		cparams2 := acmpca.GetCertificateInput{
			CertificateArn:          aws.String(*output.CertificateArn),
			CertificateAuthorityArn: aws.String(p.arn),
		}

		svc.WaitUntilCertificateIssued(&cparams2)

		output2, err2 := svc.GetCertificate(&cparams2)

		if err2 != nil {
			return nil, nil, err2
		}

		// Encode server certificate with the intermediate
		certPem := []byte(*output2.Certificate + "\n")
		chainPem := []byte(*output2.CertificateChain)

		certPem = append(certPem, chainPem...)
		return certPem, nil, nil
	*/

	return nil, nil, nil
}

// decodeCSR decodes a certificate request in PEM format and returns the
func decodeCSR(data []byte) (*x509.CertificateRequest, error) {
	/*
		block, rest := pem.Decode(data)
		if block == nil || len(rest) > 0 {
			return nil, fmt.Errorf("unexpected CSR PEM on sign request")
		}
		if block.Type != "CERTIFICATE REQUEST" {
			return nil, fmt.Errorf("PEM is not a certificate request")
		}
		csr, err := x509.ParseCertificateRequest(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate request: %v", err)
		}
		if err := csr.CheckSignature(); err != nil {
			return nil, fmt.Errorf("error checking certificate request signature: %v", err)
		}
		return csr, nil
	*/

	return nil, nil
}

// generateSubject returns the first SAN that is not 127.0.0.1 or localhost. The
// CSRs generated by the Certificate resource have always those SANs. If no SANs
// are available `awspca-issuer-certificate` will be used as a subject is always
// required.
func generateSubject(sans []string) string {
	/*
		if len(sans) == 0 {
			return "awspca-issuer-certificate"
		}
		for _, s := range sans {
			if s != "127.0.0.1" && s != "localhost" {
				return s
			}
		}
		return sans[0]
	*/

	return ""
}
