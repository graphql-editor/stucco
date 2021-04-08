package project

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/graphql-editor/stucco/pkg/providers/azure/vars"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/graphql-editor/stucco/pkg/version"
)

// Router handles tasks relating to an azure router function
type Router struct {
	Vars *vars.Vars
}

// FunctionURL returns an url from which the base zip for azure router stucco function can be downloaded from
func (r *Router) FunctionURL() (*url.URL, error) {
	v := r.Vars
	if v == nil {
		v = &vars.DefaultVars
	}
	if v.AzureFunction != "" {
		return url.Parse(v.AzureFunction)
	}
	ver := version.Version
	if strings.HasPrefix(ver, "dev-") || ver == "" {
		ver = v.Relase.DevVersion
	}
	return url.Parse("https://" + v.Relase.Host + "/" + ver + "/azure/function.zip")
}

// ZipFromReader returns a reader with router zip
func (r *Router) ZipFromReader(src io.ReaderAt, size int64, extraFiles []utils.ZipData) (rc io.ReadCloser, err error) {
	var rd io.Reader
	if len(extraFiles) == 0 {
		var br *bytes.Reader
		br, err = utils.CopyToReader(utils.ReaderAtToReader(src, size))
		if err == nil {
			rd = br
		}
	}
	if err == nil && rd == nil {
		rd, err = utils.ZipAppendFromReader(src, size, extraFiles)
	}
	if err == nil && r != nil {
		rc = ioutil.NopCloser(rd)
	}
	return
}

// Zip returns a router function as a zip reader
func (r *Router) Zip(extraFiles []utils.ZipData) (rc io.ReadCloser, err error) {
	u, err := r.FunctionURL()
	if err == nil {
		var size int64
		var rrc io.ReadCloser
		rrc, size, err = utils.LocalOrRemoteReader(u)
		if err == nil {
			defer rrc.Close()
			readerAt, ok := rrc.(io.ReaderAt)
			if !ok {
				var br *bytes.Reader
				br, err = utils.CopyToReader(rrc)
				size = br.Size()
				readerAt = br
			}
			rc, err = r.ZipFromReader(readerAt, size, extraFiles)
		}
	}
	return
}
