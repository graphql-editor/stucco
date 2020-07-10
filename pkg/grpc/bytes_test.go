package grpc_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	googlegrpc "google.golang.org/grpc"
	"k8s.io/klog"
)

func TestScanInvalid(t *testing.T) {
	data := []struct {
		input           string
		atEOF           bool
		expected        string
		expectedAdvance int
	}{
		{
			input:           string(append([]byte("return matching bytes"), 0xFF)),
			expected:        "return matching bytes",
			expectedAdvance: 22,
		},
		{
			input:           "wait for more",
			expected:        "",
			expectedAdvance: 0,
		},
		{
			input:           "return what's left atEOF",
			atEOF:           true,
			expected:        "return what's left atEOF",
			expectedAdvance: 24,
		},
	}
	for _, tt := range data {
		advance, token, err := grpc.ScanInvalid([]byte(tt.input), tt.atEOF)
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, string(token), tt)
		assert.Equal(t, tt.expectedAdvance, advance)
	}
}

type byteStreamMock struct {
	mock.Mock
	data chan []byte
	googlegrpc.ClientStream
}

func (m *byteStreamMock) Recv() (*proto.ByteStream, error) {
	m.Called()
	d, ok := <-m.data
	if !ok {
		return nil, io.EOF
	}
	return &proto.ByteStream{
		Data: d,
	}, nil
}

func TestClientLogging(t *testing.T) {
	flagSet := flag.NewFlagSet("klogflags", flag.ContinueOnError)
	klog.InitFlags(flagSet)
	flagSet.Parse([]string{"-v=5"})
	oldStderr := os.Stderr
	defer func() {
		os.Stderr = oldStderr
	}()
	pr, pw, _ := os.Pipe()
	os.Stderr = pw
	streams := map[string]*byteStreamMock{
		"Stdout": {
			data: make(chan []byte),
		},
		"Stderr": {
			data: make(chan []byte),
		},
	}
	driverClientMock := new(driverClientMock)
	for ioName, stream := range streams {
		stream.On("Recv")
		driverClientMock.
			On(ioName, mock.Anything, new(proto.ByteStreamRequest)).
			Return(stream, nil)
	}
	client := grpc.Client{
		Client: driverClientMock,
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		assert.NoError(t, client.Stdout(context.Background(), "logger"))
	}()
	go func() {
		defer wg.Done()
		assert.NoError(t, client.Stderr(context.Background(), "logger"))
	}()
	data := []struct {
		input    []byte
		expected string
		ioName   string
	}{
		{
			input:    []byte(`{"level": "info", "message": "info message"}`),
			expected: "^I.*] logger: info message",
			ioName:   "Stdout",
		},
		{
			input:    []byte(`{"level": "debug", "message": "debug message"}`),
			expected: "^I.*] logger: debug message",
			ioName:   "Stdout",
		},
		{
			input:    []byte(`[INFO]info message`),
			expected: "^I.*] logger: info message",
			ioName:   "Stdout",
		},
		{
			input:    []byte(`[DEBUG]debug message`),
			expected: "^I.*] logger: debug message",
			ioName:   "Stdout",
		},
		{
			input:    []byte(`{"level": "error", "message": "error message"}`),
			expected: "^E.*] logger: error message",
			ioName:   "Stderr",
		},
		{
			input:    []byte(`{"level": "warn", "message": "warn message"}`),
			expected: "^W.*] logger: warn message",
			ioName:   "Stderr",
		},
		{
			input:    []byte(`[ERROR]error message`),
			expected: "^E.*] logger: error message",
			ioName:   "Stderr",
		},
		{
			input:    []byte(`[WARN]warn message`),
			expected: "^W.*] logger: warn message",
			ioName:   "Stderr",
		},
	}
	t.Run("PipeLogs", func(t *testing.T) {
		for i := range data {
			tt := data[i]
			t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
				t.Parallel()
				streams[tt.ioName].data <- tt.input
			})
		}
	})
	for _, stream := range streams {
		close(stream.data)
	}
	wg.Wait()
	klog.Flush()
	pw.Close()
	b, _ := ioutil.ReadAll(pr)
	lines := strings.Split(string(b), "\n")
	lines = lines[:len(lines)-1]
	for _, tt := range data {
		assert.Condition(t, func() bool {
			expectedLines := strings.Split(tt.expected, "\n")
			re := regexp.MustCompile(expectedLines[0])
			for i := 0; i < len(lines); i++ {
				if re.Match([]byte(lines[i])) {
					forward := 1
					for j := forward; j < len(expectedLines); j++ {
						if lines[i+j] == expectedLines[j] {
							forward = j + 1
						}
					}
					if forward == len(expectedLines) {
						lines = append(lines[:i], lines[i+forward:]...)
						return true
					}
				}
			}
			return false
		})
	}
	assert.Len(t, lines, 0)
}

type mockByteStreamSend struct {
	mock.Mock
	googlegrpc.ServerStream
}

func (m *mockByteStreamSend) Send(stream *proto.ByteStream) error {
	called := m.Called(stream)
	return called.Error(0)
}

func TestIOHookHandlers(t *testing.T) {
	t.Run("Stdout", func(t *testing.T) {
		oldStdout := os.Stdout
		defer func() {
			os.Stdout = oldStdout
		}()
		pr, pw, _ := os.Pipe()
		os.Stdout = pw
		mockByteStreamSend := new(mockByteStreamSend)
		data := []string{"msg"}
		t.Run("Write", func(t *testing.T) {
			handler := grpc.PipeStdout{}
			assert.NoError(t, handler.Open())
			defer func() {
				assert.NoError(t, handler.Close())
			}()
			t.Run("HandleStreams", func(t *testing.T) {
				t.Parallel()
				assert.NoError(t, handler.Handle(new(proto.ByteStreamRequest), mockByteStreamSend))
			})
			for i := range data {
				msg := data[i]
				expectedStream := &proto.ByteStream{
					Data: []byte(msg),
				}
				mockByteStreamSend.On("Send", expectedStream).Return(nil)
				fmt.Fprint(os.Stdout, string(msg))
			}
		})
		pw.Close()
		assert.Equal(t, pw, os.Stdout)
		b, _ := ioutil.ReadAll(pr)
		for _, tt := range data {
			assert.True(t, bytes.HasPrefix(b, []byte(tt)), string(b), tt)
			b = b[len(tt):]
			mockByteStreamSend.AssertCalled(t, "Send", &proto.ByteStream{
				Data: []byte(tt),
			})
		}
	})
	t.Run("Stderr", func(t *testing.T) {
		oldStderr := os.Stderr
		defer func() {
			os.Stderr = oldStderr
		}()
		pr, pw, _ := os.Pipe()
		os.Stderr = pw
		mockByteStreamSend := new(mockByteStreamSend)
		data := []string{"msg"}
		t.Run("Write", func(t *testing.T) {
			handler := grpc.PipeStderr{}
			assert.NoError(t, handler.Open())
			defer func() {
				assert.NoError(t, handler.Close())
			}()
			t.Run("HandleStreams", func(t *testing.T) {
				t.Parallel()
				assert.NoError(t, handler.Handle(new(proto.ByteStreamRequest), mockByteStreamSend))
			})
			for i := range data {
				msg := data[i]
				expectedStream := &proto.ByteStream{
					Data: []byte(msg),
				}
				mockByteStreamSend.On("Send", expectedStream).Return(nil)
				fmt.Fprint(os.Stderr, string(msg))
			}
		})
		pw.Close()
		assert.Equal(t, pw, os.Stderr)
		b, _ := ioutil.ReadAll(pr)
		for _, tt := range data {
			assert.True(t, bytes.HasPrefix(b, []byte(tt)), string(b), tt)
			b = b[len(tt):]
			mockByteStreamSend.AssertCalled(t, "Send", &proto.ByteStream{
				Data: []byte(tt),
			})
		}
	})
}

type mockByteStreamHandlerFuncs struct {
	mock.Mock
}

func (m *mockByteStreamHandlerFuncs) Stdout(p *proto.ByteStreamRequest, d proto.Driver_StdoutServer) error {
	return m.Called(p, d).Error(0)
}

func (m *mockByteStreamHandlerFuncs) Stderr(p *proto.ByteStreamRequest, d proto.Driver_StderrServer) error {
	return m.Called(p, d).Error(0)
}

func TestHandlerFuncs(t *testing.T) {
	mockByteStreamHandlerFuncs := new(mockByteStreamHandlerFuncs)
	byteStreamRequest := new(proto.ByteStreamRequest)
	mockByteStreamHandlerFuncs.On(
		"Stdout",
		byteStreamRequest,
		proto.Driver_StdoutServer(nil),
	).Return(nil).Once()
	mockByteStreamHandlerFuncs.On("Stderr",
		byteStreamRequest,
		proto.Driver_StderrServer(nil),
	).Return(nil).Once()
	grpc.StdoutHandlerFunc(mockByteStreamHandlerFuncs.Stdout).Handle(
		byteStreamRequest,
		nil,
	)
	grpc.StderrHandlerFunc(mockByteStreamHandlerFuncs.Stderr).Handle(
		byteStreamRequest,
		nil,
	)
	mockByteStreamHandlerFuncs.AssertCalled(
		t,
		"Stdout",
		byteStreamRequest,
		(proto.Driver_StdoutServer)(nil),
	)
	mockByteStreamHandlerFuncs.AssertCalled(
		t,
		"Stderr",
		byteStreamRequest,
		proto.Driver_StdoutServer(nil),
	)
}

func TestServerCallsIOHandlers(t *testing.T) {
	mockByteStreamHandlerFuncs := new(mockByteStreamHandlerFuncs)
	byteStreamRequest := new(proto.ByteStreamRequest)
	mockByteStreamHandlerFuncs.On(
		"Stdout",
		byteStreamRequest,
		proto.Driver_StdoutServer(nil),
	).Return(nil).Once()
	mockByteStreamHandlerFuncs.On("Stderr",
		byteStreamRequest,
		proto.Driver_StderrServer(nil),
	).Return(nil).Once()
	srv := grpc.Server{}
	srv.StdoutHandler = grpc.StdoutHandlerFunc(mockByteStreamHandlerFuncs.Stdout)
	srv.StderrHandler = grpc.StderrHandlerFunc(mockByteStreamHandlerFuncs.Stderr)
	assert.NoError(t, srv.Stdout(byteStreamRequest, nil))
	assert.NoError(t, srv.Stderr(byteStreamRequest, nil))
	mockByteStreamHandlerFuncs.AssertCalled(
		t,
		"Stdout",
		byteStreamRequest,
		(proto.Driver_StdoutServer)(nil),
	)
	mockByteStreamHandlerFuncs.AssertCalled(
		t,
		"Stderr",
		byteStreamRequest,
		proto.Driver_StdoutServer(nil),
	)
}

func TestServerDoesNotRequireIOHandlers(t *testing.T) {
	byteStreamRequest := new(proto.ByteStreamRequest)
	srv := grpc.Server{}
	assert.NoError(t, srv.Stdout(byteStreamRequest, nil))
	assert.NoError(t, srv.Stderr(byteStreamRequest, nil))
}
