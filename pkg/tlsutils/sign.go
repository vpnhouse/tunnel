package tlsutils

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"

	"google.golang.org/grpc/credentials"
)

type SignGenOptions struct {
	parentSign   *Sign
	templateCert *x509.Certificate

	signer crypto.Signer
}

type Sign struct {
	cert              *x509.Certificate
	certPem           []byte
	certPrivateKey    crypto.PrivateKey
	certPrivateKeyPem []byte
}

func (s *Sign) GetCertPem() []byte {
	return s.certPem
}

func (s *Sign) GrpcServerCredentials() (credentials.TransportCredentials, error) {
	if len(s.certPem) == 0 || len(s.certPrivateKeyPem) == 0 {
		return nil, errors.New("incomplete/uninitialized cert and private key")
	}

	certificate, err := tls.X509KeyPair(s.certPem, s.certPrivateKeyPem)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS key pair: %w", err)
	}

	return credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{certificate},
	}), nil
}

func (s *Sign) GrpcClientCredentials() (credentials.TransportCredentials, error) {
	if len(s.certPem) == 0 {
		return nil, errors.New("incomplete/uninitialized cert")
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(s.certPem) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	return credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
	}), nil
}

type SignGenOption func(opts *SignGenOptions) error

func WithCA() SignGenOption {
	return func(opts *SignGenOptions) error {
		opts.templateCert.IsCA = true
		opts.templateCert.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		return nil
	}
}

func WithIPAddresses(ips ...net.IP) SignGenOption {
	return func(opts *SignGenOptions) error {
		opts.templateCert.IPAddresses = append(opts.templateCert.IPAddresses, ips...)
		return nil
	}
}

func WithLocalIPAddresses() SignGenOption {
	return func(opts *SignGenOptions) error {
		opts.templateCert.IPAddresses = append(opts.templateCert.IPAddresses, net.IPv4(127, 0, 0, 1), net.IPv6loopback)
		return nil
	}
}

func WithRsaSigner(bits int) SignGenOption {
	return func(opts *SignGenOptions) error {
		signer, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return err
		}
		opts.signer = signer
		return nil
	}
}

func WithParentSign(sign *Sign) SignGenOption {
	return func(opts *SignGenOptions) error {
		if sign == nil || sign.cert == nil || sign.certPrivateKey == nil {
			return errors.New("parent sign is invalid")
		}
		opts.parentSign = sign
		return nil
	}
}

func GenerateSign(opts ...SignGenOption) (*Sign, error) {
	serialNumMax := &big.Int{}
	serialNumMax = serialNumMax.Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumMax)
	if err != nil {
		return nil, err
	}

	var genOpts SignGenOptions

	genOpts.templateCert = &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    time.Now().Add(-10 * time.Minute).UTC(),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageKeyAgreement,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},

		UnknownExtKeyUsage:    nil,
		BasicConstraintsValid: true,
	}

	for _, opt := range opts {
		err := opt(&genOpts)
		if err != nil {
			return nil, err
		}
	}

	if genOpts.templateCert.NotAfter.IsZero() {
		// 1 year
		genOpts.templateCert.NotAfter = time.Now().AddDate(1, 0, 0)
	}

	if genOpts.signer == nil {
		return nil, errors.New("signer is not set")
	}

	var targetPublicKey crypto.PublicKey
	var signerCert *x509.Certificate
	var privateKey crypto.PrivateKey
	if genOpts.templateCert.IsCA {
		signerCert = genOpts.templateCert
		privateKey = genOpts.signer
	} else if genOpts.parentSign != nil {
		signerCert = genOpts.parentSign.cert
		if genOpts.parentSign.cert.NotAfter.Before(genOpts.templateCert.NotAfter) {
			genOpts.templateCert.NotAfter = genOpts.parentSign.cert.NotAfter
		}
		privateKey = genOpts.parentSign.certPrivateKey
	} else {
		return nil, errors.New("incomplete sign options")
	}

	genOpts.templateCert.SubjectKeyId, err = generateSubjectKeyID(targetPublicKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, genOpts.templateCert, signerCert, genOpts.signer.Public(), privateKey)
	if err != nil {
		return nil, err
	}

	certPem, err := encodePEM(cert, PemBlockTypeCertificate)
	if err != nil {
		return nil, err
	}

	certPrivateKey, err := marshalPrivateKey(genOpts.signer)
	if err != nil {
		return nil, err
	}

	certPrivateKeyPem, err := encodePEM(certPrivateKey, PemBlockTypePrivateKey)
	if err != nil {
		return nil, err
	}

	return &Sign{
		certPem:           certPem,
		certPrivateKeyPem: certPrivateKeyPem,
	}, nil
}

func generateSubjectKeyID(publicKey crypto.PublicKey) ([]byte, error) {
	data, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	return hash[:], nil
}

func encodePEM(data []byte, blockType PemBlockType) ([]byte, error) {
	pemBlock := &pem.Block{
		Type:  string(blockType),
		Bytes: data,
	}

	var buffer bytes.Buffer
	err := pem.Encode(&buffer, pemBlock)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func marshalPrivateKey(privateKey crypto.PrivateKey) ([]byte, error) {
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return rsaPrivateKeyMarshal(k)
	default:
		return nil, fmt.Errorf("unsupported private key type: (%v) %T", privateKey, privateKey)
	}
}

func rsaPrivateKeyMarshal(privateKey *rsa.PrivateKey) ([]byte, error) {
	key, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return key, nil
}
