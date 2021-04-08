package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// IsLocal returns true if url scheme is empty or equal file
func IsLocal(u *url.URL) bool {
	return u.Scheme == "" || u.Scheme == "file"
}

func fileWithSize(u *url.URL) (f *os.File, size int64, err error) {
	f, err = os.Open(u.Path)
	if err == nil {
		var fi os.FileInfo
		fi, err = os.Stat(u.Path)
		if err == nil {
			size = fi.Size()
		}
	}
	return
}

func fetchWithSize(u *url.URL) (rc io.ReadCloser, size int64, err error) {
	var resp *http.Response
	resp, err = http.Get(u.String())
	if err == nil {
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			err = fmt.Errorf("could not fetch %s, returned with error code %d", u.String(), resp.StatusCode)
		}
	}
	if err == nil {
		rc = resp.Body
		size = resp.ContentLength
	}
	return
}

// LocalOrRemoteReader creates a closable reader from url
func LocalOrRemoteReader(u *url.URL) (rc io.ReadCloser, size int64, err error) {
	if IsLocal(u) {
		rc, size, err = fileWithSize(u)
	} else {
		rc, size, err = fetchWithSize(u)
	}
	return
}

// ReadLocalOrRemoteFile loads file from local storage or http depending on scheme in url
func ReadLocalOrRemoteFile(fn string) (b []byte, err error) {
	var u *url.URL
	u, err = url.Parse(fn)
	if err == nil {
		var rc io.ReadCloser
		rc, _, err = LocalOrRemoteReader(u)
		if err == nil {
			defer rc.Close()
			b, err = ioutil.ReadAll(rc)
		}
	}
	return
}
