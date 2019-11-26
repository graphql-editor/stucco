package plugin

import (
	"bytes"
	"strings"
	"sync"
	"unsafe"

	"github.com/hashicorp/go-hclog"
	"k8s.io/klog"
)

var (
	_trimLevel = map[hclog.Level]int{
		hclog.Debug: len("[DEBUG]"),
		hclog.Trace: len("[TRACE]"),
		hclog.Info:  len("[INFO] "),
		hclog.Warn:  len("[WARN] "),
		hclog.Error: len("[ERROR]"),
	}
)

type klogLogger struct {
	hclog.Logger
	bytes.Buffer
	writeLock      sync.Mutex
	trimTimeFormat int
	namedLock      sync.Mutex
	namedLoggers   map[string]hclog.Logger
	name           string
}

func white(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r'
}

func (k *klogLogger) cleanLine(line string, lv hclog.Level) string {
	return line[k.trimTimeFormat+_trimLevel[lv]+2:]
}

func (k *klogLogger) lines(lv hclog.Level) (string, []string) {
	bs := k.Bytes()
	lines := strings.Split(*(*string)(unsafe.Pointer(&bs)), "\n")
	for i := 0; i < len(lines); i++ {
		for j := len(lines[i]) - 1; j >= 0 && white(lines[i][j]); j-- {
			lines[i] = lines[i][:j]
		}
		if len(lines[i]) == 0 {
			lines = append(lines[:i], lines[i+1:]...)
		} else {
			lines[i] = k.cleanLine(lines[i], lv)
		}
	}
	return lines[0], lines[1:]
}

func (k *klogLogger) Trace(msg string, args ...interface{}) {
	k.Logger.Trace(msg)
	k.writeLock.Lock()
	defer k.writeLock.Unlock()
	defer k.Reset()
	first, rest := k.lines(hclog.Trace)
	klog.Error(first)
	for _, l := range rest {
		klog.Error(l)
	}
}

func (k *klogLogger) Debug(msg string, args ...interface{}) {
	if klog.V(5) {
		k.Logger.Debug(msg)
		k.writeLock.Lock()
		defer k.writeLock.Unlock()
		defer k.Reset()
		first, rest := k.lines(hclog.Debug)
		klog.V(5).Info(first)
		for _, l := range rest {
			klog.V(5).Info(l)
		}
	}
}

func (k *klogLogger) Info(msg string, args ...interface{}) {
	if klog.V(3) {
		k.writeLock.Lock()
		defer k.writeLock.Unlock()
		defer k.Reset()
		k.Logger.Info(msg)
		first, rest := k.lines(hclog.Info)
		klog.V(3).Info(first)
		for _, l := range rest {
			klog.V(3).Info(l)
		}
	}
}

func (k *klogLogger) Warn(msg string, args ...interface{}) {
	k.Logger.Warn(msg)
	k.writeLock.Lock()
	defer k.writeLock.Unlock()
	defer k.Reset()
	first, rest := k.lines(hclog.Warn)
	klog.Warning(first)
	for _, l := range rest {
		klog.Warning(l)
	}
}

func (k *klogLogger) Error(msg string, args ...interface{}) {
	k.Logger.Error(msg)
	k.writeLock.Lock()
	defer k.writeLock.Unlock()
	defer k.Reset()
	first, rest := k.lines(hclog.Error)
	klog.Error(first)
	for _, l := range rest {
		klog.Error(l)
	}
}

func (k *klogLogger) Named(name string) hclog.Logger {
	k.namedLock.Lock()
	defer k.namedLock.Unlock()
	if _, ok := k.namedLoggers[name]; !ok {
		k.namedLoggers[name] = newLogger(k.name + "." + name)
	}
	return k.namedLoggers[name]
}

// replace hclog with klog for consistency
func newLogger(name string) hclog.Logger {
	l := &klogLogger{
		trimTimeFormat: len(hclog.TimeFormat),
		name:           name,
		namedLoggers:   make(map[string]hclog.Logger),
	}
	l.Logger = hclog.New(&hclog.LoggerOptions{
		Output: l,
		// doesn't matter, we check verbosity levels on klog
		// before writting either way, so always Debug is fine
		Level: hclog.Debug,
		Mutex: &l.writeLock,
		Name:  l.name,
	})
	return l
}
