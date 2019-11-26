package grpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/graphql-editor/stucco/pkg/proto"
	"k8s.io/klog"
)

type byteStream interface {
	Recv() (*proto.ByteStream, error)
}

func streamBytes(dst io.WriteCloser, stream byteStream) {
	defer dst.Close()
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			klog.Error(err)
			return
		}
		dst.Write(resp.GetData())
	}
}

func (m *GRPCClient) Stdout(name string) error {
	stream, err := m.client.Stdout(context.Background(), new(proto.ByteStreamRequest))
	if err != nil {
		return err
	}
	pr, pw := io.Pipe()
	name = name + ": "
	go func() {
		scanner := bufio.NewScanner(pr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			t := scanner.Text()
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(t), &m); err == nil {
				msg := fmt.Sprintf("%s%v", name, m["message"])
				switch m["level"] {
				case "info":
					klog.V(3).Info(msg)
				case "debug":
					fallthrough
				default:
					klog.V(5).Info(msg)
				}
			} else {
				switch {
				case strings.HasPrefix(t, "[INFO]"):
					klog.V(3).Info(name + t[len("[INFO]"):])
				case strings.HasPrefix(t, "[DEBUG]"):
					t = t[len("[DEBUG]"):]
					fallthrough
				default:
					klog.V(5).Info(name + t)
				}
			}
		}
	}()
	go streamBytes(pw, stream)
	return nil
}

func (m *GRPCClient) Stderr(name string) error {
	stream, err := m.client.Stderr(context.Background(), new(proto.ByteStreamRequest))
	if err != nil {
		return err
	}
	pr, pw := io.Pipe()
	name = name + ": "
	go func() {
		scanner := bufio.NewScanner(pr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			t := scanner.Text()
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(t), &m); err == nil {
				msg := fmt.Sprintf("%s%v", name, m["message"])
				switch m["level"] {
				case "warn":
					klog.Warning(msg)
				case "error":
					fallthrough
				default:
					klog.Error(msg)
				}
			} else {
				switch {
				case strings.HasPrefix(t, "[WARN]"):
					klog.Warning(name + t[len("[WARN]"):])
				case strings.HasPrefix(t, "[ERROR]"):
					t = t[len("[ERROR]"):]
					fallthrough
				default:
					klog.Error(name + t)
				}
			}
		}
	}()
	go streamBytes(pw, stream)
	return nil
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

type StdoutHandlerFunc func(*proto.ByteStreamRequest, proto.Driver_StdoutServer) error

type pipeStdoutCloser struct {
	old  *os.File
	pipe *os.File
}

func (p pipeStdoutCloser) Close() error {
	os.Stdout = p.old
	return p.pipe.Close()
}

func PipeStdout() (StdoutHandlerFunc, io.Closer, error) {
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	oldStdout := os.Stdout
	os.Stdout = pw
	return func(s *proto.ByteStreamRequest, ss proto.Driver_StdoutServer) error {
		gw := byteStreamWriter{ss}
		go func() {
			r := io.TeeReader(pr, oldStdout)
			io.Copy(gw, r)
		}()
		return nil
	}, pipeStdoutCloser{oldStdout, pw}, nil
}

type StderrHandlerFunc func(*proto.ByteStreamRequest, proto.Driver_StderrServer) error

type pipeStderrCloser struct {
	old  *os.File
	pipe *os.File
}

func (p pipeStderrCloser) Close() error {
	os.Stderr = p.old
	return p.pipe.Close()
}

func PipeStderr() (StderrHandlerFunc, io.Closer, error) {
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	oldStderr := os.Stderr
	os.Stderr = pw
	return func(s *proto.ByteStreamRequest, ss proto.Driver_StderrServer) error {
		gw := byteStreamWriter{ss}
		go func() {
			r := io.TeeReader(pr, oldStderr)
			io.Copy(gw, r)
		}()
		return nil
	}, pipeStderrCloser{oldStderr, pw}, nil
}

func (m *GRPCServer) Stdout(s *proto.ByteStreamRequest, ss proto.Driver_StdoutServer) error {
	if m.StdoutHandler != nil {
		return m.StdoutHandler(s, ss)
	}
	return nil
}

func (m *GRPCServer) Stderr(s *proto.ByteStreamRequest, ss proto.Driver_StderrServer) error {
	if m.StderrHandler != nil {
		return m.StderrHandler(s, ss)
	}
	return nil
}
