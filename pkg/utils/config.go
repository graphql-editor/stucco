package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type decodeFunc func([]byte, interface{}) error

func yamlUnmarshal(b []byte, v interface{}) error {
	return yaml.Unmarshal(b, v)
}

var supportedExtension = map[string]decodeFunc{
	".json": json.Unmarshal,
	".yaml": yamlUnmarshal,
	".yml":  yamlUnmarshal,
}

func getConfigExt(fn string) (ext string, err error) {
	u, err := url.Parse(fn)
	if err != nil {
		return
	}
	for k := range supportedExtension {
		if strings.HasSuffix(u.Path, k) {
			return k, nil
		}
	}
	if u.Scheme != "" {
		return "", fmt.Errorf("remote config path must be end with extension")
	}
	var st os.FileInfo
	for k := range supportedExtension {
		st, err = os.Stat(fn + k)
		if err == nil || !os.IsNotExist(err) {
			ext = k
			break
		}
	}
	if err != nil || st.IsDir() {
		if os.IsNotExist(err) {
			err = fmt.Errorf("could not find stucco config in current directory")
		}
		if err == nil {
			err = fmt.Errorf("%s is a directory", st.Name())
		}
	}
	return
}

func realConfigFileName(fn string) (configPath string, err error) {
	ext, err := getConfigExt(fn)
	if err == nil {
		configPath = fn
		if !strings.HasSuffix(fn, ext) {
			configPath = fn + ext
		}
	}
	return
}

// ReadConfigFile loads stucco config from json or yaml file.
//
// If extension is provided function loads config directly, otherwise it tries .json, .yaml and .yml extensions.
func ReadConfigFile(fn string) (b []byte, err error) {
	configPath, err := realConfigFileName(fn)
	if err == nil {
		u, err := url.Parse(configPath)
		if err == nil {
			if u.Scheme == "" {
				b, err = ioutil.ReadFile(u.Path)
			} else {
				var resp *http.Response
				resp, err = http.Get(u.String())
				if err == nil {
					defer resp.Body.Close()
					b, err = ioutil.ReadAll(resp.Body)
				}
			}
		}
	}
	return
}

// LoadConfigFile returns Config from file
func LoadConfigFile(fn string, v interface{}) (err error) {
	configPath, err := realConfigFileName(fn)
	var b []byte
	if err == nil {
		b, err = ReadConfigFile(configPath)
	}
	if err == nil {
		ext := configPath[strings.LastIndex(configPath, "."):]
		decode := supportedExtension[ext]
		err = decode(b, v)
	}
	return
}
