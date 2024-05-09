package certController

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"os/exec"
	"time"
)

type CertController struct {
	time       time.Duration
	PrivateKey *rsa.PrivateKey
	CaCert     *x509.Certificate
	WebCert    *x509.Certificate
}

func New() *CertController {
	cm := &CertController{time: 3650 * 24 * time.Hour}
	cm.generatePrivateKey(2048)
	return cm
}

func (cm *CertController) generatePrivateKey(bits int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	cm.PrivateKey = privateKey
	return nil
}

func (cm *CertController) GenerateCA() error {
	// 证书时间
	notBefore := time.Now()
	notAfter := notBefore.Add(cm.time)

	// 生成证书序列号
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{""},
			Province:           []string{""},
			Locality:           []string{""},
			Organization:       []string{"Login Helper GO"},
			OrganizationalUnit: []string{"Login Helper GO"},
			CommonName:         "Login Helper GO",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &cm.PrivateKey.PublicKey, cm.PrivateKey)
	if err != nil {
		return err
	}

	caCert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return err
	}
	cm.CaCert = caCert

	return nil
}

func (cm *CertController) GenerateCert(hostnames []string) error {
	notBefore := time.Now()
	notAfter := notBefore.Add(cm.time)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}
	// 生成证书
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{""},
			Province:           []string{""},
			Locality:           []string{""},
			Organization:       []string{"Login Helper GO"},
			OrganizationalUnit: []string{"Login Helper GO"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		DNSNames:              hostnames,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, cm.CaCert, &cm.PrivateKey.PublicKey, cm.PrivateKey)
	if err != nil {
		return err
	}

	caCert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return err
	}
	cm.WebCert = caCert

	return nil
}

func (cm *CertController) ImportToRoot(fn string) (bool, error) {
	cmd := exec.Command("certutil", "-addstore", "-f", "Root", fn)
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (cm *CertController) ExportKey(fn string) (bool, error) {
	pemKey := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(cm.PrivateKey),
	}

	file, err := os.Create(fn)
	if err != nil {
		return false, err
	}
	defer func(file *os.File) {
		file.Close()
	}(file)

	err = pem.Encode(file, pemKey)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (cm *CertController) ExportCert(fn string, cert *x509.Certificate) (bool, error) {
	pemCert := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	file, err := os.Create(fn)
	if err != nil {
		return false, err
	}
	defer func(file *os.File) {
		file.Close()
	}(file)

	err = pem.Encode(file, pemCert)

	if err != nil {
		return false, err
	}
	return true, nil
}
