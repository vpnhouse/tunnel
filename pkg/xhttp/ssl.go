/*
 * // Copyright 2021 The VPN House Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package xhttp

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-chi/chi/v5"
	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

// certData holds the SSL certificate data
// within the renewal info for LetsEncrypt.
// The `certData` struct can be serialized
// to store the certificate on a disk.
type certData struct {
	Domain string `json:"domain"`
	// Cert is a pem-encoded certificate bytes
	Cert []byte `json:"cert"`
	// key is a pem-encoded certificate key
	Key []byte `json:"key"`
	// URL is a letsEncryptRenew url for a certificate
	URL string `json:"url"`

	// cert is a parsed tls certificate
	// with .Leaf filled.
	cert tls.Certificate
}

// parseX509 loads Cert and Key bytes into x509 certificate
// with the leaf fields. Have to be called after issuing certificate from
// LE or loading the cached one from a disk.
func (data *certData) parseX509() error {
	// convert certificate for stdlib's tls.Certificate
	incert, err := tls.X509KeyPair(data.Cert, data.Key)
	if err != nil {
		return xerror.WInternalError("ssl", "failed to parse x509 key pair", err)
	}

	// Parse extra fields, especially we want to get
	// incert.Leaf.NotAfter and incert.Leaf.NotBefore
	incert.Leaf, err = x509.ParseCertificate(incert.Certificate[0])
	if err != nil {
		return xerror.WInternalError("ssl", "failed to parse the leaf", err)
	}

	data.cert = incert
	return nil
}

// https://wiki.mozilla.org/Security/Server_Side_TLS#Intermediate_compatibility_.28recommended.29
func (data *certData) intoTLS() *tls.Config {
	if len(data.cert.Certificate) == 0 {
		panic("certificate not parsed")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{
			data.cert,
		},
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.CurveP384,
			tls.X25519,
		},
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}
}

type SSLConfig struct {
	// ListenAddr for HTTPS server, default: ":443"
	ListenAddr string `yaml:"listen_addr" valid:"listen_addr,required"`
}

// DomainConfig is the YAML version of the `adminAPI.DomainConfig` struct.
type DomainConfig struct {
	Mode     string `yaml:"mode" valid:"required"`
	Name     string `yaml:"name" valid:"dns,required"`
	IssueSSL bool   `yaml:"issue_ssl,omitempty"`
	Schema   string `yaml:"schema,omitempty"`

	// Dir to store cached certificates, use sub-directory of cfgDir if possible.
	Dir string `yaml:"dir,omitempty" valid:"path"`
}

func (c *DomainConfig) Validate() error {
	modes := map[string]struct{}{
		string(adminAPI.DomainConfigModeReverseProxy): {},
		string(adminAPI.DomainConfigModeDirect):       {},
	}
	schemas := map[string]struct{}{
		"http":  {},
		"https": {},
	}

	if len(c.Name) == 0 {
		return xerror.EInternalError("domain.name is required", nil)
	}
	if len(c.Mode) == 0 {
		return xerror.EInternalError("domain.mode is required", nil)
	}
	if _, ok := modes[c.Mode]; !ok {
		return xerror.EInternalError("domain.mode got unknown value, expecting `direct` or `reverse-proxy`", nil)
	}

	if c.Mode == string(adminAPI.DomainConfigModeReverseProxy) {
		if len(c.Schema) == 0 {
			return xerror.EInternalError("domain.schema is required for the reverse-proxy mode", nil)
		}
		if _, ok := schemas[c.Schema]; !ok {
			return xerror.EInternalError("domain.schema got unknown value, expecting `http` or `https`", nil)
		}
	}

	return nil
}

type IssuerOpts struct {
	Domain   string
	CacheDir string
	Email    string

	// Router of the http (not https!) server
	Router chi.Router
	// Restart callback fired when the valid certificate issued;
	// accepts the configuration of the new cert.
	Callback func(c *tls.Config)
}

func (opts IssuerOpts) cachedCertPath() string {
	return filepath.Join(opts.CacheDir, opts.Domain+".json")
}

// Issuer should be named like certManager
type Issuer struct {
	opts IssuerOpts

	// log is a named logger with a domain name
	log *zap.Logger

	// cli is a cached LE client
	// with a valid registration field(s).
	cli *lego.Client
}

func NewIssuer(opts IssuerOpts) (*Issuer, error) {
	if opts.Callback == nil {
		return nil, errors.New("no callback is given")
	}
	if opts.Router == nil {
		return nil, errors.New("no http router is given")
	}

	fi, err := os.Stat(opts.CacheDir)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New(opts.CacheDir + ": not a directory")
	}
	if len(opts.Email) == 0 {
		opts.Email = "noreply@dummy.org"
	}

	return &Issuer{
		opts: opts,
		log:  zap.L().With(zap.String("domain", opts.Domain)),
	}, nil
}

// TLSConfig issues new certificate via LE or loads one from a cache, if any.
// Returns TLS config for the http.Server using the issued cert.
func (is *Issuer) TLSConfig() (*tls.Config, error) {
	certData, err := is.loadCachedCertificate()
	if err == nil {
		zap.L().Debug("successfully loaded cached certificate")
		go is.letsEncryptRenewWorker(certData, true)
		return certData.intoTLS(), nil
	}

	certData, err = is.issueAndSaveCertificate()
	if err == nil {
		zap.L().Debug("successfully issued new certificate")
		go is.letsEncryptRenewWorker(certData, false)
		return certData.intoTLS(), nil
	}

	go func() {
		certData := is.retryIssue()
		is.opts.Callback(certData.intoTLS())
		is.letsEncryptRenewWorker(certData, false)
	}()

	selfSigned, err := is.selfSigned()
	if err != nil {
		return nil, err
	}

	return selfSigned.intoTLS(), nil
}

func (is *Issuer) loadCachedCertificate() (certData, error) {
	cached := is.opts.cachedCertPath()
	bs, err := os.ReadFile(cached)
	if err != nil {
		// do not wrap error, we'll check for os.IsNotExists
		return certData{}, err
	}

	var cdata certData
	if err := json.Unmarshal(bs, &cdata); err != nil {
		return certData{}, err
	}

	if err := cdata.parseX509(); err != nil {
		return certData{}, err
	}

	is.log.Debug("using an existing certificate", zap.String("path", cached))
	return cdata, nil
}

func (is *Issuer) issueAndSaveCertificate() (certData, error) {
	cdata, err := is.issue()
	if err != nil {
		return certData{}, err
	}

	if err := is.writeCert(cdata); err != nil {
		return certData{}, err
	}

	return cdata, nil
}

func (is *Issuer) writeCert(cdata certData) error {
	bs, err := json.Marshal(cdata)
	if err != nil {
		return xerror.WInternalError("ssl", "failed to marshal cert data", err)
	}

	dataFile := is.opts.cachedCertPath()
	if err := os.WriteFile(dataFile, bs, 0600); err != nil {
		return xerror.WInternalError("ssl", "failed to write cert data to a file",
			err, zap.String("path", dataFile))
	}

	is.log.Debug("certificate saved", zap.String("path", dataFile))
	return nil
}

func (is *Issuer) getLEClient() (*lego.Client, error) {
	if is.cli == nil {
		myUser := newUser(is.opts.Email)

		config := lego.NewConfig(&myUser)
		// A client facilitates communication with the CA server.
		client, err := lego.NewClient(config)
		if err != nil {
			return nil, xerror.WInternalError("ssl", "failed to create LE client", err)
		}

		h01p := newHttp01provider(is.opts.Router)
		if err := client.Challenge.SetHTTP01Provider(h01p); err != nil {
			return nil, xerror.WInternalError("ssl", "failed to set http-01 provider", err)
		}

		// register new user
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return nil, xerror.WInternalError("ssl", "failed to register user", err)
		}

		myUser.registration = reg
		is.cli = client
	}

	return is.cli, nil
}

func (is *Issuer) issue() (certData, error) {
	cli, err := is.getLEClient()
	if err != nil {
		return certData{}, err
	}

	request := certificate.ObtainRequest{
		Domains: []string{is.opts.Domain},
		Bundle:  true,
	}
	certificates, err := cli.Certificate.Obtain(request)
	if err != nil {
		return certData{}, xerror.WInternalError("ssl", "failed to obtain certificate", err)
	}

	cd := certData{
		Domain: is.opts.Domain,
		Cert:   certificates.Certificate,
		Key:    certificates.PrivateKey,
		URL:    certificates.CertURL,
	}

	if err := cd.parseX509(); err != nil {
		return certData{}, err
	}

	is.log.Info("certificate issued")
	return cd, nil
}

func (is *Issuer) selfSigned() (certData, error) {
	is.log.Warn("using self-signed certificate")
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return certData{}, xerror.WInternalError("ssl", "failed to generate ecdsa key", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return certData{}, xerror.WInternalError("ssl", "failed to randomize serial number", err)
	}

	template := x509.Certificate{
		DNSNames: []string{is.opts.Domain},
		// IPAddresses    []net.IP // << TODO?
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"VPN House"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return certData{}, xerror.WInternalError("ssl", "failed to create certificate from a template", err)
	}

	// Create public key
	pubBuf := new(bytes.Buffer)
	err = pem.Encode(pubBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return certData{}, xerror.WInternalError("ssl", "failed to encode certificate into PEM", err)
	}

	// Create private key
	privBuf := new(bytes.Buffer)
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return certData{}, xerror.WInternalError("ssl", "failed to marshal the private key", err)
	}

	err = pem.Encode(privBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return certData{}, xerror.WInternalError("ssl", "failed to encode the private key", err)
	}

	cert, err := tls.X509KeyPair(pubBuf.Bytes(), privBuf.Bytes())
	if err != nil {
		return certData{}, xerror.WInternalError("ssl", "failed to derive x509 certificate", err)
	}

	cdata := certData{
		Domain: is.opts.Domain,
		cert:   cert,
	}

	is.log.Debug("using self-signed certificate")
	return cdata, nil
}

func (is *Issuer) retryIssue() certData {
	is.log.Debug("failed to issue the certificate, retrying")

	t := time.NewTimer(time.Minute)
	defer t.Stop()

	backoff := 1
	for {
		select {
		case <-t.C:
			cdata, err := is.issueAndSaveCertificate()
			if err == nil {
				return cdata
			}

			next := timepow(backoff)
			// handle overflow
			backoff++
			if backoff > 8 {
				backoff = 8
			}

			t.Reset(next)
		}
	}
}

// letsEncryptRenew tracks the LE renewal process.
// The following timings are applied here:
// - every 24 hours if expiration >= 60 days but < 89 days;
// - backoff timer with range of 1...256 minutes if the cert is valid for less than 24 hours;
// - same backoff used to retry in case of errors from LE;
func (is *Issuer) letsEncryptRenewWorker(cdata certData, immediate bool) {
	next := time.Duration(0)
	if !immediate {
		next = 24 * time.Hour
	}

	t := time.NewTimer(next)
	defer t.Stop()

	backoff := 1
	for {
		select {
		case <-t.C:
			next = 24 * time.Hour
			if ok := is.renewOnce(cdata); !ok {
				next = timepow(backoff)
				backoff++
				if backoff > 8 {
					backoff = 8
				}
			}

			t.Reset(next)
		}
	}
}

// renewOnce attempts to renew the existing certificate via the LE
func (is *Issuer) renewOnce(existing certData) bool {
	const t30days = 30 * 24 * time.Hour
	expiresIn := existing.cert.Leaf.NotAfter.Sub(time.Now())
	more30daysToExpire := expiresIn > t30days
	if more30daysToExpire {
		is.log.Debug("more than 30 days to expire, nothing to do", zap.Duration("expire_in", expiresIn))
		return true
	}

	client, err := is.getLEClient()
	if err != nil {
		return false
	}

	newCert, err := client.Certificate.Renew(certificate.Resource{
		Domain:      existing.Domain,
		CertURL:     existing.URL,
		PrivateKey:  existing.Key,
		Certificate: existing.Cert,
	}, true, false, "")
	if err != nil {
		is.log.Warn("renew failed", zap.Error(err))
		return false
	}

	cdata := certData{
		Domain: existing.Domain,
		Cert:   newCert.Certificate,
		Key:    newCert.PrivateKey,
		URL:    newCert.CertURL,
	}
	if err := cdata.parseX509(); err != nil {
		return false
	}

	if err := is.writeCert(cdata); err != nil {
		return false
	}

	is.opts.Callback(cdata.intoTLS())
	return true
}

func timepow(backoff int) time.Duration {
	v := 1 << backoff
	return time.Duration(v) * time.Minute
}
