package protohttp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/golang/protobuf/proto"
	protobuf "github.com/golang/protobuf/proto"
)

// Client implements driver by using Protocol Buffers over HTTP
type Client struct {
	*http.Client
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
		Client: config.Client,
		URL:    config.URL,
	}
}

type message struct {
	contentType protobufMessageContentType
	proto       proto.Message
}

func (c *Client) do(in, out message) error {
	b, err := proto.Marshal(in.proto)
	if err == nil {
		var resp *http.Response
		resp, err = c.Post(c.URL, in.contentType.String(), bytes.NewReader(b))
		if err == nil {
			defer resp.Body.Close()
			err = unmarshalFromHTTP(resp, out)
		}
	}
	return err
}

func contentTypesEqual(a, b string) bool {
	return a == b
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

func unmarshalFromHTTP(
	resp *http.Response,
	out message,
) error {
	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = fmt.Errorf(string(b))
		}
		return err
	}
	messageType, err := getMessageType(resp.Header.Get(contentTypeHeader))
	if err != nil {
		return err
	}
	if string(out.contentType) != messageType {
		return fmt.Errorf("cannot unmarshal %s to %s", messageType, string(out.contentType))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		err = protobuf.Unmarshal(body, out.proto)
	}
	return err
}
