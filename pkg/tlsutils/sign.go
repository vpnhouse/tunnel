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
	"os"
	"path"
	"time"

	"google.golang.org/grpc/credentials"
)

const (
	certFileName       = "tls.cert"
	privateKeyFileName = "tls.key"
)

var ErrStorageDir = errors.New("storage directory is empty")

func makeName(signName string, fileName string) string {
	if signName == "" {
		return fileName
	}
	return fmt.Sprintf("%s-%s", signName, fileName)
}

type SignGenOptions struct {
	parentSign   *Sign
	templateCert *x509.Certificate
	signer       crypto.Signer
}

type Sign struct {
	Cert          *x509.Certificate
	CertPem       []byte
	PrivateKey    crypto.PrivateKey
	PrivateKeyPem []byte
}

func (s *Sign) String() string {
	if s.Cert != nil {
		return fmt.Sprintf(
			"cert nb: %s, na: %s",
			s.Cert.NotBefore.Format(time.RFC3339),
			s.Cert.NotAfter.Format(time.RFC3339),
		)
	}
	return ""
}

func (s *Sign) GrpcServerCredentials() (credentials.TransportCredentials, error) {
	if len(s.CertPem) == 0 || len(s.PrivateKeyPem) == 0 {
		return nil, errors.New("incomplete/uninitialized cert and private key")
	}

	certificate, err := tls.X509KeyPair(s.CertPem, s.PrivateKeyPem)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS key pair: %w", err)
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
	}), nil
}

func (s *Sign) GrpcClientCredentials() (credentials.TransportCredentials, error) {
	if len(s.CertPem) == 0 {
		return nil, errors.New("incomplete/uninitialized cert")
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(s.CertPem) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	return credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
	}), nil
}

func (s *Sign) Store(storageDirectory string, signName string) error {
	if storageDirectory == "" {
		return ErrStorageDir
	}
	if len(s.CertPem) == 0 {
		return errors.New("sign cert is not set on")
	}
	if len(s.PrivateKeyPem) == 0 {
		return errors.New("sign private key is not set on")
	}

	err := os.WriteFile(path.Join(storageDirectory, makeName(signName, certFileName)), s.CertPem, 0600)
	if err != nil {
		return fmt.Errorf("failed to store sign cert: %w", err)
	}

	err = os.WriteFile(path.Join(storageDirectory, makeName(signName, privateKeyFileName)), s.PrivateKeyPem, 0600)
	if err != nil {
		return fmt.Errorf("failed to store sign cert: %w", err)
	}

	return nil
}

func LoadSign(storageDirectory string, signName string) (*Sign, error) {
	if storageDirectory == "" {
		return nil, ErrStorageDir
	}
	certPem, err := os.ReadFile(path.Join(storageDirectory, makeName(signName, certFileName)))
	if err != nil {
		return nil, fmt.Errorf("failed to load sign cert: %w", err)
	}

	crtDer, err := decodePEM(certPem, PemBlockTypeCertificate)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(crtDer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sign cert: %w", err)
	}

	privateKeyPem, err := os.ReadFile(path.Join(storageDirectory, makeName(signName, privateKeyFileName)))
	if err != nil {
		return nil, fmt.Errorf("failed to load sign private key: %w", err)
	}

	privateKeyDer, err := decodePEM(privateKeyPem, PemBlockTypePrivateKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := unmarshalPrivateKey(privateKeyDer)
	if err != nil {
		return nil, err
	}

	return &Sign{
		Cert:          cert,
		CertPem:       certPem,
		PrivateKey:    privateKey,
		PrivateKeyPem: privateKeyPem,
	}, nil

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
		if sign == nil || sign.Cert == nil || sign.PrivateKey == nil {
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
		SerialNumber:          serialNumber,
		NotBefore:             time.Now().Add(-10 * time.Minute).UTC(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, opt := range opts {
		err := opt(&genOpts)
		if err != nil {
			return nil, err
		}
	}

	if genOpts.signer == nil {
		return nil, errors.New("signer is not set")
	}

	var signerCert *x509.Certificate
	var privateKey crypto.PrivateKey
	var publicKey crypto.PublicKey
	if genOpts.templateCert.IsCA {
		signerCert = genOpts.templateCert
		privateKey = genOpts.signer
		publicKey = genOpts.signer.Public()
	} else if genOpts.parentSign != nil {
		signerCert = genOpts.parentSign.Cert
		privateKey = genOpts.parentSign.PrivateKey
		publicKey, err = getPublicKey(genOpts.signer)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("incomplete sign options")
	}

	genOpts.templateCert.SubjectKeyId, err = generateSubjectKeyID(publicKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, genOpts.templateCert, signerCert, publicKey, privateKey)
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
		Cert:          genOpts.templateCert,
		CertPem:       certPem,
		PrivateKey:    genOpts.signer,
		PrivateKeyPem: certPrivateKeyPem,
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

func decodePEM(data []byte, blockType PemBlockType) ([]byte, error) {
	pemBlock, _ := pem.Decode(data)
	if pemBlock == nil {
		return nil, errors.New("cannot parse PEM block")
	}

	if pemBlock.Type != string(blockType) {
		return nil, fmt.Errorf("unexpected pem block %s, expected %s", pemBlock.Type, blockType)
	}

	if len(pemBlock.Headers) != 0 {
		return nil, errors.New("invalid PEM block, no headers")
	}

	return pemBlock.Bytes, nil
}

func marshalPrivateKey(privateKey crypto.PrivateKey) ([]byte, error) {
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		key, err := x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			return nil, err
		}
		return key, nil
	default:
		return nil, fmt.Errorf("unsupported private key type: (%v) %T", privateKey, privateKey)
	}
}

func unmarshalPrivateKey(keyDer []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(keyDer); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(keyDer)
	if err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("unknown private key type (%v) %T in PKCS#8 wrapping", key, key)
		}
	}

	return nil, errors.New("unsupported private key type")
}

func getPublicKey(pivateKey crypto.PrivateKey) (crypto.PublicKey, error) {
	type privateToPublicKey interface {
		Public() crypto.PublicKey
	}
	if p, ok := pivateKey.(privateToPublicKey); ok {
		return p.Public(), nil
	}
	return nil, fmt.Errorf("unsupported type of private key: %v", pivateKey)
}
