/*
 * // Copyright 2021 The VPN House Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package xhttp

import (
	"bytes"
	"context"
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
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-chi/chi/v5"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

// TODO: Add graceful shutdown

// leCertInfo holds the SSL certificate data
// within the renewal info for LetsEncrypt.
// The `leCertInfo` struct can be serialized
// to store the certificate on a disk.
type leCertInfo struct {
	Domain string `json:"domain"`
	// Cert is a pem-encoded certificate bytes
	Cert []byte `json:"cert"`
	// key is a pem-encoded certificate key
	Key []byte `json:"key"`
	// URL is a letsEncryptRenew url for a certificate
	URL string `json:"url"`
}

type SSLConfig struct {
	// ListenAddr for HTTPS server, default: ":443"
	ListenAddr string `yaml:"listen_addr" valid:"listen_addr,required"`
}

type IssuerOpts struct {
	Domain       string
	CacheDir     string
	Email        string
	NonSSLRouter chi.Router
}

func (opts IssuerOpts) cachedCertPath() string {
	return filepath.Join(opts.CacheDir, opts.Domain+".json")
}

// Issuer should be named like certManager
type Issuer struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts *IssuerOpts

	// log is a named logger with a domain name
	log *zap.Logger

	// cli is a cached LE client
	// with a valid registration field(s).
	cli *lego.Client

	// Let's Encrypt certificate information
	leCert *leCertInfo

	// Current parsed TLS certificate
	tlsCert *tls.Certificate
}

func NewIssuer(opts *IssuerOpts) (*Issuer, error) {
	if opts.NonSSLRouter == nil {
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

	ctx, cancel := context.WithCancel(context.Background())
	issuer := &Issuer{
		ctx:    ctx,
		cancel: cancel,
		opts:   opts,
		log:    zap.L().With(zap.String("domain", opts.Domain)),
	}

	// We actually should ignore error here relying on re-issue certificate
	err = issuer.loadCachedCertificate()
	if err == nil {
		zap.L().Debug("successfully loaded cached certificate")
	} else {
		err = issuer.assignSelfSigned()
		if err != nil {
			return nil, err
		}
	}

	go issuer.worker()

	return issuer, nil
}

func (is *Issuer) GetCertificate(domain string) *tls.Certificate {
	if strings.EqualFold(domain, is.opts.Domain) {
		return is.tlsCert
	}

	return nil
}

func (is *Issuer) workerTask(task func() error) {
	minInterval := time.Minute
	maxInterval := time.Hour

	interval := minInterval
	for {
		err := task()
		if err == nil {
			break
		}

		select {
		case <-is.ctx.Done():
			return
		case <-time.After(interval):
		}

		// LE has limit of 5 failures per hour, we will start with 1, 4, 16, 60 minutes and continue once per hour
		interval *= 4
		if interval > maxInterval {
			interval = maxInterval
		}
	}
}

func (is *Issuer) worker() {
	if is.leCert == nil {
		is.workerTask(is.issueAndSaveCertificate)
	}
	is.workerTask(is.renewOnce)

	for {
		select {
		case <-is.ctx.Done():
			return
		case <-time.After(time.Duration(24) * time.Hour):
		}

		is.workerTask(is.renewOnce)
	}
}

func (is *Issuer) Shutdown() {
	is.cancel()
}

func (is *Issuer) loadCachedCertificate() error {
	cached := is.opts.cachedCertPath()
	bs, err := os.ReadFile(cached)
	if err != nil {
		return err
	}

	var leCert leCertInfo
	err = json.Unmarshal(bs, &leCert)
	if err != nil {
		return err
	}

	tlsCert, err := parseX509(leCert.Cert, leCert.Key)
	if err != nil {
		return err
	}

	is.leCert = &leCert
	is.tlsCert = tlsCert

	is.log.Debug("using an existing certificate", zap.String("path", cached))
	return nil
}

func (is *Issuer) writeCert() error {
	bs, err := json.Marshal(is.leCert)
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

		h01p := newHttp01provider(is.opts.NonSSLRouter)
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

func (is *Issuer) issueAndSaveCertificate() error {
	cli, err := is.getLEClient()
	if err != nil {
		return err
	}

	request := certificate.ObtainRequest{
		Domains: []string{is.opts.Domain},
		Bundle:  true,
	}
	certificates, err := cli.Certificate.Obtain(request)
	if err != nil {
		return xerror.WInternalError("ssl", "failed to obtain certificate", err)
	}

	leCert := leCertInfo{
		Domain: is.opts.Domain,
		Cert:   certificates.Certificate,
		Key:    certificates.PrivateKey,
		URL:    certificates.CertURL,
	}

	tlsCert, err := parseX509(leCert.Cert, leCert.Key)
	if err != nil {
		return err
	}

	is.leCert = &leCert
	is.tlsCert = tlsCert
	is.log.Info("new SSL certificate issued")

	if err := is.writeCert(); err != nil {
		return err
	}

	is.log.Info("certificate successfully stored")
	return nil
}

func (is *Issuer) assignSelfSigned() error {
	is.log.Warn("using self-signed certificate")
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return xerror.WInternalError("ssl", "failed to generate ecdsa key", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return xerror.WInternalError("ssl", "failed to randomize serial number", err)
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
		return xerror.WInternalError("ssl", "failed to create certificate from a template", err)
	}

	// Create public key
	pubBuf := new(bytes.Buffer)
	err = pem.Encode(pubBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return xerror.WInternalError("ssl", "failed to encode certificate into PEM", err)
	}

	// Create private key
	privBuf := new(bytes.Buffer)
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return xerror.WInternalError("ssl", "failed to marshal the private key", err)
	}

	err = pem.Encode(privBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return xerror.WInternalError("ssl", "failed to encode the private key", err)
	}

	cert, err := tls.X509KeyPair(pubBuf.Bytes(), privBuf.Bytes())
	if err != nil {
		return xerror.WInternalError("ssl", "failed to derive x509 certificate", err)
	}
	is.tlsCert = &cert

	is.log.Debug("using self-signed certificate")
	return nil
}

// renewOnce attempts to renew the existing certificate via the LE
func (is *Issuer) renewOnce() error {
	if is.leCert == nil || is.tlsCert == nil {
		return xerror.WInternalError("ssl", "Can't renew, certificate is not set", nil)
	}

	const t30days = 30 * 24 * time.Hour
	expiresIn := time.Until(is.tlsCert.Leaf.NotAfter)
	more30daysToExpire := expiresIn > t30days
	if more30daysToExpire {
		is.log.Debug("more than 30 days to expire, nothing to do", zap.Duration("expire_in", expiresIn))
		return nil
	}

	client, err := is.getLEClient()
	if err != nil {
		return err
	}

	newCert, err := client.Certificate.Renew(certificate.Resource{
		Domain:      is.leCert.Domain,
		CertURL:     is.leCert.URL,
		PrivateKey:  is.leCert.Key,
		Certificate: is.leCert.Cert,
	}, true, false, "")
	if err != nil {
		return xerror.WInternalError("ssl", "Renew failed", err)
	}

	tlsCert, err := parseX509(newCert.Certificate, newCert.PrivateKey)
	if err != nil {
		return err
	}

	is.leCert = &leCertInfo{
		Domain: is.leCert.Domain,
		Cert:   newCert.Certificate,
		Key:    newCert.PrivateKey,
		URL:    newCert.CertURL,
	}
	is.tlsCert = tlsCert

	if err := is.writeCert(); err != nil {
		return err
	}

	return nil
}

// parseX509 loads Cert and Key bytes into x509 certificate
// with the leaf fields. Have to be called after issuing certificate from
// LE or loading the cached one from a disk.
func parseX509(cert, key []byte) (*tls.Certificate, error) {
	// convert certificate for stdlib's tls.Certificate
	incert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, xerror.WInternalError("ssl", "failed to parse x509 key pair", err)
	}

	// Parse extra fields, especially we want to get
	// incert.Leaf.NotAfter and incert.Leaf.NotBefore
	incert.Leaf, err = x509.ParseCertificate(incert.Certificate[0])
	if err != nil {
		return nil, xerror.WInternalError("ssl", "failed to parse the leaf", err)
	}

	return &incert, nil
}
