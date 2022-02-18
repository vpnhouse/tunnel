/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
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

	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-chi/chi/v5"
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
	// Domain name to issue the certificate for,
	// self-signed certificate is used if name is empty but ssl has been enabled,
	// or if external issuer failed.
	Domain string `yaml:"domain" valid:"dns"`
	// Email to notify about certificates expiration, optional.
	Email string `yaml:"email" valid:"email"`
	// ListenAddr for HTTPS server, default: ":443"
	ListenAddr string `yaml:"listen_addr" valid:"listen_addr,required"`
	// Dir to store cached certificates, use sub-directory of cfgDir if possible.
	Dir string `yaml:"dir" valid:"path"`
}

// Issuer should be named like certManager
type Issuer struct {
	// router of the http (not https!) server
	router chi.Router
	log    *zap.Logger

	// cli is a cached LE client
	// with a valid registration field(s).
	cli *lego.Client

	cfg             SSLConfig
	restartCallback func(newCert *tls.Config)
}

func NewIssuer(cfg SSLConfig, r chi.Router, cb func(c *tls.Config)) (*Issuer, error) {
	fi, err := os.Stat(cfg.Dir)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New(cfg.Dir + ": not a directory")
	}
	if len(cfg.Email) == 0 {
		cfg.Email = "noreply@dummy.org"
	}

	return &Issuer{
		router:          r,
		cfg:             cfg,
		restartCallback: cb,
		log:             zap.L().With(zap.String("domain", cfg.Domain)),
	}, nil
}

// TLSConfig issues new certificate via LE or loads one from a cache, if any.
// Returns TLS config for the http.Server using the issued cert.
func (is *Issuer) TLSConfig() (*tls.Config, error) {
	certData, err := is.loadCachedCertificate()
	if err == nil {
		return certData.intoTLS(), nil
	}

	certData, err = is.issueAndSaveCertificate()
	if err == nil {
		go is.letsEncryptRenew(certData)
		return certData.intoTLS(), nil
	}

	go func() {
		certData := is.retryIssue()
		is.restartCallback(certData.intoTLS())
		is.letsEncryptRenew(certData)
	}()

	selfSigned, err := is.selfSigned()
	if err != nil {
		return nil, err
	}

	return selfSigned.intoTLS(), nil
}

func (is *Issuer) loadCachedCertificate() (certData, error) {
	cached := is.cachedCertPath()
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

func (is *Issuer) cachedCertPath() string {
	return filepath.Join(is.cfg.Dir, is.cfg.Domain+".json")
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

	dataFile := is.cachedCertPath()
	if err := os.WriteFile(dataFile, bs, 0600); err != nil {
		return xerror.WInternalError("ssl", "failed to write cert data to a file",
			err, zap.String("path", dataFile))
	}

	is.log.Debug("certificate saved", zap.String("path", dataFile))
	return nil
}

func (is *Issuer) getLEClient() (*lego.Client, error) {
	if is.cli == nil {
		myUser := newUser(is.cfg.Email)

		config := lego.NewConfig(&myUser)
		// A client facilitates communication with the CA server.
		client, err := lego.NewClient(config)
		if err != nil {
			return nil, xerror.WInternalError("ssl", "failed to create LE client", err)
		}

		h01p := newHttp01provider(is.router)
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
		Domains: []string{is.cfg.Domain},
		Bundle:  true,
	}
	certificates, err := cli.Certificate.Obtain(request)
	if err != nil {
		return certData{}, xerror.WInternalError("ssl", "failed to obtain certificate", err)
	}

	cd := certData{
		Domain: is.cfg.Domain,
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
		DNSNames: []string{is.cfg.Domain},
		// IPAddresses    []net.IP // << TODO?
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Uranium VPN"},
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
		Domain: is.cfg.Domain,
		cert:   cert,
	}

	is.log.Debug("using self-signed certificate")
	return cdata, nil
}

func (is *Issuer) retryIssue() certData {
	is.log.Debug("failed to issue the certificate, retrying")

	t := time.NewTicker(3 * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			cdata, err := is.issueAndSaveCertificate()
			if err == nil {
				return cdata
			}
		}
	}
}

func (is *Issuer) letsEncryptRenew(cdata certData) {
	t := time.NewTimer(10 * time.Minute)
	defer t.Stop()

	previousOK := true
	for {
		select {
		case <-t.C:
			previousOK = is.renewOnce(cdata, previousOK)
		}
	}
}

func (is *Issuer) renewOnce(existing certData, previousOK bool) bool {
	expiresSoon := existing.cert.Leaf.NotAfter.Sub(time.Now()) <= 24*time.Hour
	if previousOK && !expiresSoon {
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

	is.restartCallback(cdata.intoTLS())
	return true
}
