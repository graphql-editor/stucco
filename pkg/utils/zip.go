package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"
)

// ZipData Is a tuple with data and filename to be added to zip archive
type ZipData struct {
	Data     io.Reader
	Filename string
	Info     os.FileInfo
	Symlink  bool
}

// AddFileToZip adds file represented by ZipData to zip archive
func (z *ZipData) AddFileToZip(zipWriter *zip.Writer) error {
	var header *zip.FileHeader
	if z.Info != nil {
		var err error
		header, err = zip.FileInfoHeader(z.Info)
		if err != nil {
			return err
		}
	}
	if header == nil {
		header = &zip.FileHeader{
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		var m os.FileMode
		m = 0755
		if z.Symlink {
			m &= os.ModeSymlink
		}
		header.SetMode(0755)
	}

	header.Name = z.Filename
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, z.Data)
	return err
}

// ZipFiles create new zip archive with files.
func ZipFiles(filename string, files []ZipData) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err = file.AddFileToZip(zipWriter); err != nil {
			return err
		}
	}
	return nil
}

type zipFile struct {
	header zip.FileHeader
	data   io.ReadCloser
}

func getFile(f *zip.File) (zf zipFile, err error) {
	zf.header = f.FileHeader
	r, err := f.Open()
	if err != nil {
		return
	}
	zf.data = r
	return
}

func getFiles(zipReader *zip.Reader) (files []zipFile, err error) {
	for _, f := range zipReader.File {
		var zf zipFile
		zf, err = getFile(f)
		if err != nil {
			return
		}
		files = append(files, zf)
	}
	return
}

// ZipAppend appends data to existing zip archive
func ZipAppend(filename string, files []ZipData) error {
	ff, err := os.Open(filename)
	if err != nil {
		return err
	}
	fi, err := os.Stat(filename)
	if err != nil {
		ff.Close()
		return err
	}
	defer ff.Close()
	r, err := ZipAppendFromReader(ff, fi.Size(), files)
	if err == nil {
		var f *os.File
		f, err = os.OpenFile(filename, os.O_WRONLY, fi.Mode().Perm())
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, r)
	}
	return err
}

// ZipAppendFromReader appends data to existing zip archive overwritting files that match name
func ZipAppendFromReader(r io.ReaderAt, size int64, files []ZipData) (io.Reader, error) {
	f, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	currentFiles, err := getFiles(f)
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, cf := range currentFiles {
			cf.data.Close()
		}
	}()
	n := len(files)
	exists := func(f zipFile) bool {
		return sort.Search(n, func(i int) bool {
			return files[i].Filename == f.header.Name
		}) != n
	}
	// Overwrite existing files
	for _, cf := range currentFiles {
		if !exists(cf) {
			files = append(files, ZipData{
				Filename: cf.header.Name,
				Data:     cf.data,
				Info:     cf.header.FileInfo(),
			})
		}
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	for _, file := range files {
		if err := file.AddFileToZip(zipWriter); err != nil {
			return nil, err
		}
	}
	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}

// AddPathToZip appends contents of directory to zip, optionaly skipping files/directories matching gitignore like lines
func AddPathToZip(path string, ignoreGlobs []string, w *zip.Writer) error {
	return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		cp, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}
		if cp == "." {
			if !d.IsDir() {
				return fmt.Errorf("%s must be a directory", path)
			}
			return nil
		}
		gitIgnore := ignore.CompileIgnoreLines(ignoreGlobs...)
		if gitIgnore.MatchesPath(cp) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			fi, err := d.Info()
			if err != nil {
				return err
			}
			var dc io.ReadCloser
			if fi.Mode()&os.ModeSymlink != 0 {
				target, err := os.Readlink(filepath.Join(path, cp))
				if err != nil {
					return err
				}
				pathAbs, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				pabs, err := filepath.Abs(filepath.Dir(filepath.Join(path, cp)))
				if err != nil {
					return err
				}
				r := filepath.Clean(filepath.Join(pabs, target))
				if err != nil {
					return err
				}
				tabs, err := filepath.Abs(r)
				if err != nil {
					return err
				}
				if !strings.HasPrefix(tabs, pathAbs) {
					return fmt.Errorf("link %s target (%s) is outside of archive", cp, target)
				}
				dc = io.NopCloser(strings.NewReader(target))
			} else {
				var f *os.File
				f, err = os.Open(p)
				if err != nil {
					return err
				}
				dc = f
			}
			zd := ZipData{
				Filename: cp,
				Info:     fi,
				Data:     dc,
			}
			err = zd.AddFileToZip(w)
			dc.Close()
		}
		return err
	})
}
