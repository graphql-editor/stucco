package grpc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/graphql-editor/stucco/pkg/proto"
	"k8s.io/klog"
)

type byteStream interface {
	Recv() (*proto.ByteStream, error)
}

// ScanInvalid implements bufio.SplitFunc for bufio.Scanner, it reads up to 0xFF (invalid unicode)
// or EOF and returns read data without 0xFF.
func ScanInvalid(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0xFF); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func streamBytes(dst io.WriteCloser, stream byteStream) error {
	defer dst.Close()
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		data := resp.GetData()
		data = append(data, 0xFF)
		dst.Write(data)
	}
}

// Stdout opens a byte stream between from server to client and logs results using k8s.io/klog
//
// Checks if byte stream is valid json with property level set to info or debug and logs the contents of
// message property from json. Otherwise if byte stream has prefix [INFO] or [DEBUG] logs the contents of stream to
// matching verbosity level without prefix. If not matched, logs the whole byte stream unmodified with debug verbosity.
//
// Info verbosity is 3
// Debug verbosity is 5
func (m *Client) Stdout(ctx context.Context, name string) error {
	stream, err := m.Client.Stdout(ctx, new(proto.ByteStreamRequest))
	if err != nil {
		return err
	}
	pr, pw := io.Pipe()
	errCh := make(chan error)
	go func() {
		errCh <- streamBytes(pw, stream)
	}()
	name = name + ": "
	scanner := bufio.NewScanner(pr)
	scanner.Split(ScanInvalid)
	for scanner.Scan() {
		t := scanner.Text()
		var m map[string]interface{}
		verbosity := klog.Level(5)
		if err := json.Unmarshal([]byte(t), &m); err == nil {
			switch m["level"] {
			case "info":
				verbosity = klog.Level(3)
				fallthrough
			case "debug":
				t = fmt.Sprintf("%v", m["message"])
			}
		} else {
			strip := len("[DEBUG]")
			switch {
			case strings.HasPrefix(t, "[INFO]"):
				verbosity = klog.Level(3)
				strip = len("[INFO]")
				fallthrough
			case strings.HasPrefix(t, "[DEBUG]"):
				t = t[strip:]
			}
		}
		klog.V(verbosity).Info(name + t)
	}
	return <-errCh
}

// Stderr opens a byte stream between from server to client and logs results using k8s.io/klog
//
// Checks if byte stream is valid json with property level set to warn or err and logs the contents of
// message property from json. Otherwise if byte stream has prefix [WARN] or [ERROR] logs the contents of stream to
// matching klog Severity. If not matched, logs the whole byte stream unmodified to Error severity.
func (m *Client) Stderr(ctx context.Context, name string) error {
	stream, err := m.Client.Stderr(ctx, new(proto.ByteStreamRequest))
	if err != nil {
		return err
	}
	pr, pw := io.Pipe()
	errCh := make(chan error)
	go func() {
		errCh <- streamBytes(pw, stream)
	}()
	name = name + ": "
	scanner := bufio.NewScanner(pr)
	scanner.Split(ScanInvalid)
	for scanner.Scan() {
		t := scanner.Text()
		var m map[string]interface{}
		logF := klog.Error
		if err := json.Unmarshal([]byte(t), &m); err == nil {
			switch m["level"] {
			case "warn":
				logF = klog.Warning
				fallthrough
			case "error":
				t = fmt.Sprintf("%v", m["message"])
			}
		} else {
			strip := len("[ERROR]")
			switch {
			case strings.HasPrefix(t, "[WARN]"):
				logF = klog.Warning
				strip = len("[WARN]")
				fallthrough
			case strings.HasPrefix(t, "[ERROR]"):
				t = t[strip:]
			}
		}
		logF(name + t)
	}
	return <-errCh
}

type byteStreamServer interface {
	Send(*proto.ByteStream) error
}

type byteStreamWriter struct {
	byteStreamServer
}

func (b byteStreamWriter) Write(p []byte) (int, error) {
	err := b.Send(&proto.ByteStream{
		Data: p,
	})
	n := len(p)
	if err != nil {
		n = 0
	}
	return n, err
}

// StdoutHandlerFunc is a type of function that must be implemented in server implementation
// to handle ByteStreamRequest for stdout.
type StdoutHandlerFunc func(*proto.ByteStreamRequest, proto.Driver_StdoutServer) error

// Handle for implementing StdoutHandler interface
func (f StdoutHandlerFunc) Handle(p *proto.ByteStreamRequest, d proto.Driver_StdoutServer) error {
	return f(p, d)
}

type stdoutHandlerCloser struct {
	io.Closer
	StdoutHandler
}

// HookFile implements StdoutHandler by hijacking os.Stdout, and writing all
// data to both os.Stdout and ByteStream.
type hookFile struct {
	pr, pw  *os.File
	oldFile *os.File
	lock    sync.Mutex
}

func (h *hookFile) open(f *os.File) (*os.File, error) {
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	h.pr, h.pw = pr, pw
	h.oldFile = f
	return h.pw, nil
}

func (h *hookFile) handle(s *proto.ByteStreamRequest, stream byteStreamServer) error {
	h.lock.Lock()
	pr := h.pr
	oldFile := h.oldFile
	h.lock.Unlock()
	gw := byteStreamWriter{stream}
	r := io.TeeReader(pr, oldFile)
	_, err := io.Copy(gw, r)
	return err
}

func (h *hookFile) close() (*os.File, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	var err error
	if h.pw != nil {
		err = h.pw.Close()
	}
	return h.oldFile, err
}

// PipeStdout adds a hook to os.Stdout, watching all writes done to os.Stdout.
// Writes both written to os.Stdout and to GRPC ByteStream.
type PipeStdout struct {
	hookFile
}

// Open creates a new hook on os.Stdout. Must be called before Handle.
func (p *PipeStdout) Open() error {
	f, err := p.open(os.Stdout)
	if err == nil {
		os.Stdout = f
	}
	return err
}

// Handle is a blocking method that sends all Writes to os.Stdout through GRPC
func (p *PipeStdout) Handle(req *proto.ByteStreamRequest, srv proto.Driver_StdoutServer) error {
	return p.hookFile.handle(req, srv)
}

// Close cleans up after a hook. It is an error to not close a hook.
func (p *PipeStdout) Close() error {
	f, err := p.hookFile.close()
	if f != nil {
		os.Stdout = f
	}
	return err
}

// StderrHandlerFunc is a type of function that must be implemented in server implementation
// to handle ByteStreamRequest for stderr.
type StderrHandlerFunc func(*proto.ByteStreamRequest, proto.Driver_StderrServer) error

// Handle for implementing StderrHandler interface
func (f StderrHandlerFunc) Handle(p *proto.ByteStreamRequest, d proto.Driver_StderrServer) error {
	return f(p, d)
}

// PipeStderr adds a hook to os.Stderr, watching all writes done to os.Stderr.
// Writes both written to os.Stderr and to GRPC ByteStream.
type PipeStderr struct {
	hookFile
}

// Open creates a new hook on os.Stderr. Must be called before Handle.
func (p *PipeStderr) Open() error {
	f, err := p.open(os.Stderr)
	if err == nil {
		os.Stderr = f
	}
	return err
}

// Handle is a blocking method that sends all Writes to os.Stderr through GRPC
func (p *PipeStderr) Handle(req *proto.ByteStreamRequest, srv proto.Driver_StderrServer) error {
	return p.hookFile.handle(req, srv)
}

// Close cleans up after a hook. It is an error to not close a hook.
func (p *PipeStderr) Close() error {
	f, err := p.hookFile.close()
	if f != nil {
		os.Stderr = f
	}
	return err
}

// Stdout handles a stdout bytestream request for server
func (m *Server) Stdout(s *proto.ByteStreamRequest, ss proto.Driver_StdoutServer) error {
	if m.StdoutHandler != nil {
		return m.StdoutHandler.Handle(s, ss)
	}
	return nil
}

// Stderr handles a stderr bytestream request for server
func (m *Server) Stderr(s *proto.ByteStreamRequest, ss proto.Driver_StderrServer) error {
	if m.StderrHandler != nil {
		return m.StderrHandler.Handle(s, ss)
	}
	return nil
}
