package security

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/graphql-editor/stucco/pkg/utils"
)

// CertAuth is a certificate auth for function
type CertAuth struct {
	Cert         []byte
	Key          []byte
	Renegotation tls.RenegotiationSupport
}

const (
	pemStart     = "-----BEGIN "
	pemEnd       = "-----END "
	pemEndOfLine = "-----"
)

func checkPem(data []string) bool {
	if data[0] == "" {
		data = data[1:]
	}
	last := len(data) - 1
	// Strip trailing empty data
	for data[last] == "" {
		data = data[:last]
		last = len(data) - 1
	}
	// Valid PEM starts with empty line, skip it, to allow data with and without empty line
	return strings.HasPrefix(data[0], pemStart) &&
		strings.HasSuffix(data[0], pemEndOfLine) &&
		strings.HasPrefix(data[last], pemEnd) &&
		strings.HasSuffix(data[last], pemEndOfLine)
}

func loadPem(data string) ([]byte, error) {
	lines := strings.Split(data, "\n")
	if len(lines) == 1 {
		d, err := utils.ReadLocalOrRemoteFile(data)
		if err != nil {
			return nil, err
		}
		data = string(d)
		lines = strings.Split(data, "\n")
	}
	if !checkPem(lines) {
		return nil, errors.New("invalid pem data")
	}
	return []byte(data), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (c *CertAuth) UnmarshalJSON(b []byte) error {
	cert := struct {
		Cert string `json:"cert"`
		Key  string `json:"key"`
	}{}
	if err := json.Unmarshal(b, &cert); err != nil {
		return err
	}
	c.Cert = []byte(cert.Cert)
	c.Key = []byte(cert.Key)
	return nil
}

// RoundTripper returns round tripper for cert auth
func (c CertAuth) RoundTripper() (http.RoundTripper, error) {
	cert, err := loadPem(string(c.Cert))
	if err != nil {
		return nil, err
	}
	key, err := loadPem(string(c.Key))
	if err != nil {
		return nil, err
	}
	clientCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:  []tls.Certificate{clientCert},
			Renegotiation: c.Renegotation,
		},
	}, nil
}

// KeyAuth simple key based auth for function
type KeyAuth struct {
	Key  string
	Next http.RoundTripper
}

// UnmarshalJSON implements json.Unmarshaler
func (k *KeyAuth) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		k.Key = s
		return nil
	}
	kk := struct {
		Key string `json:"key"`
	}{}
	if err := json.Unmarshal(b, &kk); err != nil {
		return err
	}
	k.Key = kk.Key
	return nil
}

// RoundTrip implements http.RoundTripper
func (k KeyAuth) RoundTrip(r *http.Request) (*http.Response, error) {
	rt := k.Next
	if rt == nil {
		rt = http.DefaultTransport
	}
	r.Header.Add("X-Stucco-Auth-Key", string(k.Key))
	return rt.RoundTrip(r)
}

// RoundTripper returns round tripper for key auth
func (k KeyAuth) RoundTripper() (http.RoundTripper, error) {
	return k, nil
}

// Auth is a function auth
type Auth struct {
	Key  *KeyAuth
	Cert *CertAuth

	rt http.RoundTripper
}

// UnmarshalJSON implements json.Unmarshaler
func (a *Auth) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		a.Key = &KeyAuth{Key: s}
		return nil
	}
	keyAuth := struct {
		Key *KeyAuth `json:"apiKey,omitempty"`
	}{}
	if err := json.Unmarshal(b, &keyAuth); err == nil && keyAuth.Key != nil {
		a.Key = keyAuth.Key
		return nil
	}
	var certAuth CertAuth
	if err := json.Unmarshal(b, &certAuth); err == nil && certAuth.Cert != nil {
		a.Cert = &certAuth
		return nil
	}
	return errors.New("could not unmarshal function auth data")
}

// RoundTripper wraps user provided round tripper for function with auth
func (a *Auth) RoundTripper() (rt http.RoundTripper, err error) {
	if a.rt != nil {
		return a.rt, nil
	}
	if a.Key == nil && a.Cert == nil {
		return nil, errors.New("missing auth implementation")
	}
	if a.Key != nil && a.Cert != nil {
		return nil, errors.New("only one auth implementation can be set at once")
	}
	if a.Key != nil {
		rt, err = a.Key.RoundTripper()
	}
	if a.Cert != nil {
		rt, err = a.Cert.RoundTripper()
	}
	if err == nil {
		a.rt = rt
	}
	return rt, err
}

// Security represents security configuration for function
type Security struct {
	Auth
	Functions map[string]Auth
}

// RoundTripper returns roundtripper for function
func (s Security) RoundTripper(fn string) (http.RoundTripper, error) {
	if f, ok := s.Functions[fn]; ok {
		return f.RoundTripper()
	}
	return s.Auth.RoundTripper()
}
