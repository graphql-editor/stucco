package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

// HostJSONKey is a key instance in host.json
type HostJSONKey struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Encrypted bool   `json:"encrypted"`
}

func (h HostJSONKey) decryptKey(decryptionKeyID string) (string, error) {
	return "", errors.New("not implemented")
}

// Key returns key, optionaly decrypting it if it's encrypted
func (h HostJSONKey) Key(decryptionKeyID string) (string, error) {
	if h.Encrypted {
		return h.decryptKey(decryptionKeyID)
	}
	return h.Value, nil
}

// HostJSON represents HostJSON used with azure function deployment
type HostJSON struct {
	MasterKey       HostJSONKey   `json:"masterKey"`
	FunctionKeys    []HostJSONKey `json:"functionKeys"`
	DecryptionKeyID string        `json:"decryptionKeyId,omitempty"`
}

// HasKey returns true if HostJSON has either master key or some function key
func (h HostJSON) HasKey() bool {
	return h.MasterKey.Value != "" || len(h.FunctionKeys) > 0
}

// StorageHostKeyReader read function key from azure storage host.json
type StorageHostKeyReader struct {
	Account, Key     string
	ConnectionString string
	lock             sync.RWMutex
	cache            *HostJSON
}

func (s *StorageHostKeyReader) appName(function string) (name string, err error) {
	u, err := envFuncURL(function)
	if err == nil {
		name = strings.Split(u.Hostname(), ".")[0]
	}
	return
}

func (s *StorageHostKeyReader) unsafeAccount() string {
	account := s.Account
	if account == "" {
		csParts := strings.Split(s.ConnectionString, ";")
		for _, p := range csParts {
			if strings.HasPrefix(p, "AccountName=") {
				account = strings.TrimPrefix(p, "AccountName=")
			}
		}
	}
	return account
}

func (s *StorageHostKeyReader) unsafeKey() string {
	key := s.Key
	if key == "" {
		csParts := strings.Split(s.ConnectionString, ";")
		for _, p := range csParts {
			if strings.HasPrefix(p, "AccountKey=") {
				key = strings.TrimPrefix(p, "AccountKey=")
			}
		}
	}
	return key
}

func (s *StorageHostKeyReader) unsafeCreds() (credential *azblob.SharedKeyCredential, err error) {
	account := s.unsafeAccount()
	key := s.unsafeKey()
	err = errors.New("missing azure storage account credentials")
	if account != "" && key != "" {
		credential, err = azblob.NewSharedKeyCredential(account, key)
	}
	return
}

func (s *StorageHostKeyReader) unsafeFetchCache(function string) error {
	credential, err := s.unsafeCreds()
	if err != nil {
		return err
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/azure-webjobs-secrets", s.unsafeAccount()))
	if err != nil {
		return err
	}
	appName, err := s.appName(function)
	if err != nil {
		return err
	}
	containerURL := azblob.NewContainerURL(*u, p)
	blobURL := containerURL.NewBlockBlobURL(fmt.Sprintf("%s/host.json", appName))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	get, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false)
	cancel()
	if err != nil {
		return err
	}
	responseBody := get.Body(azblob.RetryReaderOptions{})
	var h HostJSON
	if err = json.NewDecoder(responseBody).Decode(&h); err == nil {
		s.cache = &h
	}
	return err
}

func (s *StorageHostKeyReader) hostJSON(function string) (h HostJSON, err error) {
	s.lock.RLock()
	if s.cache != nil {
		h = *s.cache
	}
	s.lock.RUnlock()
	if !h.HasKey() {
		s.lock.Lock()
		defer s.lock.Unlock()
		if s.cache == nil {
			err = s.unsafeFetchCache(function)
		}
		if err == nil {
			h = *s.cache
		}
	}
	return
}

func (s *StorageHostKeyReader) getKey(function string) (k string, err error) {
	h, err := s.hostJSON(function)
	if err != nil || !h.HasKey() {
		if err == nil {
			err = errors.New("no function key found")
		}
		return
	}
	if h.MasterKey.Value != "" {
		return h.MasterKey.Key(h.DecryptionKeyID)
	}
	for _, fk := range h.FunctionKeys {
		if fk.Value != "" {
			return fk.Key(h.DecryptionKeyID)
		}
	}
	err = errors.New("no function key found")
	return
}

// GetKey returns function key from host.json
func (s *StorageHostKeyReader) GetKey(function string) (k string, err error) {
	k, err = s.getKey(function)
	if err != nil || k == "" {
		return envKeyReader{}.GetKey(function)
	}
	return
}
