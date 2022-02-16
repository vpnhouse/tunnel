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

// Issuer should be named like certManager
type Issuer struct {
	// path to a certificate cache dir
	certdir string
	// router of the http (not https!) server
	router chi.Router
	// email for LetsEncrypt registration
	email string
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

// TLSForDomain issues new certificate via LE or loads one from a cache, if any.
// Returns TLS config for the http.Server.
func (is *Issuer) TLSForDomain(domain string) (*tls.Config, error) {
	var cert tls.Certificate
	if len(domain) > 0 {
		certData, err := is.getCertificate(domain)
		if err != nil {
			return nil, err
		}
		// TODO(nikonov): start the renew routine here
		cert = certData.cert
	} else {
		var err error
		cert, err = is.selfSigned()
		if err != nil {
			return nil, err
		}
	}

	config := &tls.Config{
		// MinVersion: tls.VersionTLS12,
		Certificates: []tls.Certificate{
			cert,
		},
		// https://wiki.mozilla.org/Security/Server_Side_TLS#Intermediate_compatibility_.28recommended.29
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

	return config, nil
}

func (is *Issuer) certCachePath(dom string) string {
	return filepath.Join(is.certdir, dom+".json")
}

func (is *Issuer) getCertificate(domain string) (CertData, error) {
	cert, err := is.loadOrObtainCertificate(domain)
	if err != nil {
		return CertData{}, err
	}

	// convert certificate for stdlib's tls.Certificate
	incert, err := tls.X509KeyPair(cert.Certificate, cert.PrivateKey)
	if err != nil {
		return CertData{}, err
	}

	// Parse extra fields, especially we want to get
	// incert.Leaf.NotAfter and incert.Leaf.NotBefore
	incert.Leaf, err = x509.ParseCertificate(incert.Certificate[0])
	if err != nil {
		return CertData{}, err
	}

	return CertData{
		Domain: domain,
		Cert:   cert.Certificate,
		Key:    cert.PrivateKey,
		URL:    cert.CertURL,
		cert:   incert,
	}, nil
}

func (is *Issuer) loadOrObtainCertificate(dom string) (*certificate.Resource, error) {
	cached := is.certCachePath(dom)
	bs, err := os.ReadFile(cached)
	if err != nil {
		// not in cache, go issue the new one
		if errors.Is(err, os.ErrNotExist) {
			return is.obtainAndSave(dom)
		}
		return nil, err
	}

	var cdata CertData
	if err := json.Unmarshal(bs, &cdata); err != nil {
		return nil, err
	}

	zap.L().Debug("using an existing certificate", zap.String("domain", dom))
	return &certificate.Resource{
		Domain:      dom,
		PrivateKey:  cdata.Key,
		Certificate: cdata.Cert,
		CertURL:     cdata.URL,
	}, nil
}

func (is *Issuer) obtainAndSave(dom string) (*certificate.Resource, error) {
	zap.L().Debug("issuing certificate", zap.String("domain", dom))
	cert, err := is.obtain(dom)
	if err != nil {
		return nil, err
	}

	dataFile := is.certCachePath(dom)
	cdata := CertData{
		Domain: dom,
		Key:    cert.PrivateKey,
		Cert:   cert.Certificate,
		URL:    cert.CertURL,
	}

	bs, err := json.Marshal(cdata)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(dataFile, bs, 0600); err != nil {
		return nil, err
	}

	zap.L().Debug("certificate saved", zap.String("domain", dom), zap.String("path", dataFile))
	return cert, nil
}

func (is *Issuer) obtain(dom string) (*certificate.Resource, error) {
	myUser := newUser(is.email)

	config := lego.NewConfig(&myUser)
	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	h01p := newHttp01provider(is.router)
	if err := client.Challenge.SetHTTP01Provider(h01p); err != nil {
		return nil, err
	}

	// register new user
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	myUser.registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{dom},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}

	return certificates, nil
}

func (is *Issuer) selfSigned() (tls.Certificate, error) {
	zap.L().Warn("using self-signed certificate")
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
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
		return tls.Certificate{}, err
	}

	// Create public key
	pubBuf := new(bytes.Buffer)
	err = pem.Encode(pubBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return tls.Certificate{}, err
	}

	// Create private key
	privBuf := new(bytes.Buffer)
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	err = pem.Encode(privBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.X509KeyPair(pubBuf.Bytes(), privBuf.Bytes())
}
