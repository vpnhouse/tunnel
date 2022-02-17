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

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CertData holds the SSL certificate data
// within th renew info for LetsEncrypt.
type CertData struct {
	Domain string `json:"domain"`
	// Cert is a pem-encoded certificate bytes
	Cert []byte `json:"cert"`
	// key is a pem-encoded certificate key
	Key []byte `json:"key"`
	// URL is a renew url for a certificate
	URL string `json:"url"`

	// cert is a parsed tls certificate
	// with .Leaf filled.
	cert tls.Certificate
}

func (data *CertData) load() error {
	// convert certificate for stdlib's tls.Certificate
	incert, err := tls.X509KeyPair(data.Cert, data.Key)
	if err != nil {
		return err
	}

	// Parse extra fields, especially we want to get
	// incert.Leaf.NotAfter and incert.Leaf.NotBefore
	incert.Leaf, err = x509.ParseCertificate(incert.Certificate[0])
	if err != nil {
		return err
	}

	data.cert = incert
	return nil
}

// https://wiki.mozilla.org/Security/Server_Side_TLS#Intermediate_compatibility_.28recommended.29
func (data *CertData) intoTLS() *tls.Config {
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

// Issuer should be named like certManager
type Issuer struct {
	// path to a certificate cache dir
	certdir string
	// router of the http (not https!) server
	router chi.Router
	// email for LetsEncrypt registration
	email string

	restartCallback func(newCert *tls.Config)
}

func NewIssuer(dir string, r chi.Router) (*Issuer, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New(dir + ": not a directory")
	}

	return &Issuer{
		router:  r,
		certdir: dir,
		email:   "noreply@dummy.org",
	}, nil
}

func (is *Issuer) getConfig() {
	// try load from a disk
	// try issue from LE
	// if not LE - return self-signed, start retrying
	// if OK LE - return LE, stop retrying,	start update routine

	// implement restartCallback that accepts new tls.Config,
	// must be used for the web server restarting (probably inplace).
}

// TLSForDomain issues new certificate via LE or loads one from a cache, if any.
// Returns TLS config for the http.Server.
func (is *Issuer) TLSForDomain(domain string) (*tls.Config, error) {
	certData, err := is.loadCachedCertificate(domain)
	if err == nil {
		return certData.intoTLS(), nil
	}

	certData, err = is.issueAndSaveCertificate(domain)
	if err == nil {
		is.renew(certData)
		return certData.intoTLS(), nil
	}

	go func() {
		certData := is.retryIssue(domain)
		is.restartCallback(certData.intoTLS())
		is.renew(certData)
	}()

	selfSigned, err := is.selfSigned(domain)
	if err != nil {
		return nil, err
	}

	return selfSigned.intoTLS(), nil
}

func (is *Issuer) loadCachedCertificate(dom string) (CertData, error) {
	cached := is.certCachePath(dom)
	bs, err := os.ReadFile(cached)
	if err != nil {
		// do not wrap error, we'll check for os.IsNotExists
		return CertData{}, err
	}

	var cdata CertData
	if err := json.Unmarshal(bs, &cdata); err != nil {
		return CertData{}, err
	}

	if err := cdata.load(); err != nil {
		return CertData{}, err
	}

	zap.L().Debug("using an existing certificate", zap.String("domain", dom))
	return cdata, nil
}

func (is *Issuer) certCachePath(dom string) string {
	return filepath.Join(is.certdir, dom+".json")
}

func (is *Issuer) issueAndSaveCertificate(dom string) (CertData, error) {
	cdata, err := is.issue(dom)
	if err != nil {
		return CertData{}, err
	}

	bs, err := json.Marshal(cdata)
	if err != nil {
		return CertData{}, err
	}

	dataFile := is.certCachePath(dom)
	if err := os.WriteFile(dataFile, bs, 0600); err != nil {
		return CertData{}, err
	}

	zap.L().Debug("certificate saved", zap.String("domain", dom), zap.String("path", dataFile))
	return cdata, nil
}

func (is *Issuer) issue(dom string) (CertData, error) {
	myUser := newUser(is.email)

	config := lego.NewConfig(&myUser)
	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return CertData{}, err
	}

	h01p := newHttp01provider(is.router)
	if err := client.Challenge.SetHTTP01Provider(h01p); err != nil {
		return CertData{}, err
	}

	// register new user
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return CertData{}, err
	}
	myUser.registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{dom},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return CertData{}, err
	}

	cd := CertData{
		Domain: dom,
		Cert:   certificates.Certificate,
		Key:    certificates.PrivateKey,
		URL:    certificates.CertURL,
	}

	if err := cd.load(); err != nil {
		return CertData{}, err
	}

	zap.L().Info("certificate issued", zap.String("domain", dom))
	return cd, nil
}

func (is *Issuer) selfSigned(dom string) (CertData, error) {
	zap.L().Warn("using self-signed certificate")
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return CertData{}, err
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return CertData{}, err
	}

	template := x509.Certificate{
		DNSNames: []string{dom},
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
		return CertData{}, err
	}

	// Create public key
	pubBuf := new(bytes.Buffer)
	err = pem.Encode(pubBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return CertData{}, err
	}

	// Create private key
	privBuf := new(bytes.Buffer)
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return CertData{}, err
	}

	err = pem.Encode(privBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return CertData{}, err
	}

	cert, err := tls.X509KeyPair(pubBuf.Bytes(), privBuf.Bytes())
	if err != nil {

	}

	cdata := CertData{
		Domain: dom,
		cert:   cert,
	}

	zap.L().Debug("using self-signed certificate", zap.String("domain", dom))
	return cdata, nil
}

func (is *Issuer) retryIssue(dom string) CertData {
	t := time.NewTicker(3 * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			cdata, err := is.issueAndSaveCertificate(dom)
			if err == nil {
				return cdata
			}
		}
	}
}

func (is *Issuer) renew(cdata CertData) {
	// TODO(nikonov): ???
}
