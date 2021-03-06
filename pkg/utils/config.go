package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// StuccoConfigEnv is a name of environment variable that will be checked for stucco.json path if one is not provided
const StuccoConfigEnv = "STUCCO_CONFIG"

type decodeFunc func([]byte, interface{}) error
type encodeFunc func(interface{}) ([]byte, error)

func yamlUnmarshal(b []byte, v interface{}) error {
	return yaml.Unmarshal(b, v)
}

var supportedExtension = map[string]decodeFunc{
	".json": json.Unmarshal,
	".yaml": yamlUnmarshal,
	".yml":  yamlUnmarshal,
}
var supportedExtensionEncode = map[string]encodeFunc{
	".json": json.Marshal,
	".yaml": yaml.Marshal,
	".yml":  yaml.Marshal,
}

func getConfigExt(fn string) (ext string, isurl bool, err error) {
	u, err := url.Parse(fn)
	if err != nil {
		return
	}
	for k := range supportedExtension {
		if strings.HasSuffix(u.Path, k) {
			return k, u.Scheme != "", nil
		}
	}
	if u.Scheme != "" {
		err = errors.Errorf("remote config path must be end with extension")
		return
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
			err = errors.Errorf("could not find stucco config at %s", fn)
		}
		if err == nil {
			err = errors.Errorf("%s is a directory", st.Name())
		}
	}
	return
}

func realConfigFileName(fn string) (configPath string, err error) {
	if fn == "" {
		if env := os.Getenv(StuccoConfigEnv); env != "" {
			fn = env
		} else {
			fn = "./stucco"
		}
	}
	ext, isurl, err := getConfigExt(fn)
	if err == nil {
		configPath = fn
		if !isurl && !strings.HasSuffix(fn, ext) {
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
		return ReadLocalOrRemoteFile(configPath)
	}
	return
}

// SaveConfigFile saves config to file
func SaveConfigFile(fn string, data interface{}) (err error) {
	ext, _, _ := getConfigExt(fn)
	configPath, err := realConfigFileName(fn)
	if err != nil {
		return err
	}
	encode := supportedExtensionEncode[ext]

	file, err := encode(data)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, file, 0644)
}

// LoadConfigFile returns Config from file
func LoadConfigFile(fn string, v interface{}) (err error) {
	configPath, err := realConfigFileName(fn)
	var b []byte
	if err == nil {
		b, err = ReadConfigFile(configPath)
	}
	if err == nil {
		var u *url.URL
		u, err = url.Parse(configPath)
		if err == nil {
			// TODO: Check based on response content-type for remote configs
			ext := u.Path[strings.LastIndex(u.Path, "."):]
			decode := supportedExtension[ext]
			if decode != nil {
				err = decode(b, v)
			} else {
				err = errors.Errorf("%s is not a supported config extension", ext)
			}
		}
	}
	return
}
