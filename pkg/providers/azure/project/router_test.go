package project_test

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/graphql-editor/stucco/pkg/providers/azure/project"
	"github.com/graphql-editor/stucco/pkg/providers/azure/vars"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/graphql-editor/stucco/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noError(t *testing.T, err error) {
	require.NoError(t, err)
}

type zipFile struct {
	name string
	data []byte
}

type zipArchive struct {
	files []zipFile
}

func prepareZipData(t *testing.T, zipFn string, z zipArchive) *bytes.Reader {
	var buf bytes.Buffer
	rz := zip.NewWriter(&buf)
	for _, f := range z.files {
		h := zip.FileHeader{
			Method:   zip.Deflate,
			Name:     f.name,
			Modified: time.Now(),
		}
		h.SetMode(0755)
		w, err := rz.CreateHeader(&h)
		noError(t, err)
		_, err = w.Write(f.data)
		noError(t, err)
	}
	rz.Close()
	err := os.WriteFile(zipFn, buf.Bytes(), 0755)
	noError(t, err)
	return bytes.NewReader(buf.Bytes())
}

func readZipFiles(t *testing.T, rc io.ReadCloser) zipArchive {
	var za zipArchive
	var err error
	var br *bytes.Reader
	br, err = utils.CopyToReader(rc)
	noError(t, err)
	noError(t, rc.Close())
	var zr *zip.Reader
	zr, err = zip.NewReader(br, br.Size())
	noError(t, err)
	for _, f := range zr.File {
		var r io.ReadCloser
		r, err = f.Open()
		noError(t, err)
		var data []byte
		data, err = ioutil.ReadAll(r)
		noError(t, r.Close())
		za.files = append(za.files, zipFile{
			name: f.Name,
			data: data,
		})
	}
	return za
}

func TestRouterFunctionURL(t *testing.T) {
	var r project.Router

	version.Version = "v1.1.1"
	u, err := r.FunctionURL()
	assert.NoError(t, err)
	assert.Equal(t, u.String(), "https://stucco-release.fra1.cdn.digitaloceanspaces.com/v1.1.1/azure/function.zip")

	version.Version = ""
	u, err = r.FunctionURL()
	assert.NoError(t, err)
	assert.Equal(t, u.String(), "https://stucco-release.fra1.cdn.digitaloceanspaces.com/latest/azure/function.zip")

	r.Vars = &vars.Vars{
		AzureFunction: "/abc/def.zip",
	}
	version.Version = ""
	u, err = r.FunctionURL()
	assert.NoError(t, err)
	assert.Equal(t, u.String(), "/abc/def.zip")
}

func TestRouterZip(t *testing.T) {
	d := t.TempDir()
	routerZipFn := filepath.Join(d, "archive.zip")
	routerZip := prepareZipData(
		t,
		routerZipFn,
		zipArchive{
			files: []zipFile{
				{name: "test.txt", data: []byte("testdata")},
			},
		},
	)
	c := make(chan struct{})
	defer close(c)
	http.Handle("/", http.FileServer(http.Dir(d)))
	srv := http.Server{
		Addr: ":9999",
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	go func() {
		<-c
		srv.Shutdown(context.Background())
	}()
	td := []struct {
		extraFiles []zipFile
		expected   []zipFile
	}{
		{
			extraFiles: []zipFile{
				{
					name: "test2.txt",
					data: []byte("testdata2"),
				},
			},
			expected: []zipFile{
				{name: "test.txt", data: []byte("testdata")},
				{name: "test2.txt", data: []byte("testdata2")},
			},
		},
		{
			expected: []zipFile{
				{name: "test.txt", data: []byte("testdata")},
			},
		},
	}
	makeZipData := func(files []zipFile) (zd []utils.ZipData) {
		for _, f := range files {
			zd = append(zd, utils.ZipData{Filename: f.name, Data: bytes.NewReader(f.data)})
		}
		return
	}
	for _, tt := range td {
		check := func(t *testing.T, rc io.ReadCloser, err error) {
			assert.NoError(t, err)
			za := readZipFiles(t, rc)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, za.files)
		}
		t.Run("ZipFromReader", func(t *testing.T) {
			var r project.Router
			rc, err := r.ZipFromReader(routerZip, routerZip.Size(), makeZipData(tt.extraFiles))
			check(t, rc, err)
		})
		t.Run("Zip from file", func(t *testing.T) {
			r := project.Router{
				Vars: &vars.Vars{
					AzureFunction: routerZipFn,
				},
			}
			rc, err := r.Zip(makeZipData(tt.extraFiles))
			check(t, rc, err)
		})
		t.Run("Zip from http", func(t *testing.T) {
			r := project.Router{
				Vars: &vars.Vars{
					AzureFunction: "http://localhost:9999/archive.zip",
				},
			}
			rc, err := r.Zip(makeZipData(tt.extraFiles))
			check(t, rc, err)
		})
	}
}
