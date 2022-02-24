package protohttp

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
)

// HTTPClient for protocol buffer
type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// Client implements driver by using Protocol Buffers over HTTP
type Client struct {
	HTTPClient
	// URL of a proto server endpoint
	URL string
}

// Config for new .Client
type Config struct {
	Client *http.Client
	URL    string
}

// NewClient creates a a new client
func NewClient(config Config) Client {
	if config.Client == nil {
		config.Client = http.DefaultClient
	}
	return Client{
		HTTPClient: config.Client,
		URL:        config.URL,
	}
}

type message struct {
	contentType         protobufMessageContentType
	responseContentType protobufMessageContentType
	b                   []byte
}

func (c *Client) do(in message) ([]byte, error) {
	resp, err := c.Post(c.URL, in.contentType.String(), bytes.NewReader(in.b))
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			var b []byte
			b, err = ioutil.ReadAll(resp.Body)
			if err == nil {
				err = fmt.Errorf(`status_code=%d message="%s"`, resp.StatusCode, string(b))
			}
		}
	}
	if err == nil {
		err = in.responseContentType.checkContentType(resp.Header.Get(contentTypeHeader))
	}
	var b []byte
	if err == nil {
		b, err = ioutil.ReadAll(resp.Body)
	}
	return b, err
}

func getMessageType(contentType string) (string, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}
	if mediaType != protobufContentType {
		return "", fmt.Errorf("%s is not supported, only %s", mediaType, protobufContentType)
	}
	return params["message"], nil
}
