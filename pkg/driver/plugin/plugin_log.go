package plugin

import (
	"bytes"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"k8s.io/klog"
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

func (k *klogLogger) clean(lv hclog.Level) string {
	parts := strings.Split(k.String(), " ")[2:]
	for parts[0] == "" {
		parts = parts[1:]
	}
	return strings.Join(parts, " ")
}

// Trace writes an error log to klog
func (k *klogLogger) Trace(msg string, args ...interface{}) {
	k.writeLock.Lock()
	defer k.writeLock.Unlock()
	k.Logger.Trace(msg, args...)
	defer k.Reset()
	klog.Error(k.clean(hclog.Trace))
}

// Debug writes info log to klog with verbosity of level 5
func (k *klogLogger) Debug(msg string, args ...interface{}) {
	if klog.V(5) {
		k.writeLock.Lock()
		defer k.writeLock.Unlock()
		k.Logger.Debug(msg, args...)
		defer k.Reset()
		klog.V(5).Info(k.clean(hclog.Debug))
	}
}

// Info writes info log to klog with verbosity of level 3
func (k *klogLogger) Info(msg string, args ...interface{}) {
	if klog.V(3) {
		k.writeLock.Lock()
		defer k.writeLock.Unlock()
		k.Logger.Info(msg, args...)
		defer k.Reset()
		klog.V(3).Info(k.clean(hclog.Info))
	}
}

// Warn writes a warn log to klog
func (k *klogLogger) Warn(msg string, args ...interface{}) {
	k.writeLock.Lock()
	defer k.writeLock.Unlock()
	k.Logger.Warn(msg, args...)
	defer k.Reset()
	klog.Warning(k.clean(hclog.Info))
}

// Error writes an error log to klog
func (k *klogLogger) Error(msg string, args ...interface{}) {
	k.writeLock.Lock()
	defer k.writeLock.Unlock()
	k.Logger.Error(msg, args...)
	defer k.Reset()
	klog.Error(k.clean(hclog.Error))
}

// Named returns a name of a logger for plugin
func (k *klogLogger) Named(name string) hclog.Logger {
	k.namedLock.Lock()
	defer k.namedLock.Unlock()
	if _, ok := k.namedLoggers[name]; !ok {
		k.namedLoggers[name] = NewLogger(k.name + "." + name)
	}
	return k.namedLoggers[name]
}

// NewLogger replaces hclog used by go-plugin with klog for consistency
func NewLogger(name string) hclog.Logger {
	l := &klogLogger{
		trimTimeFormat: len(hclog.TimeFormat),
		name:           name,
		namedLoggers:   make(map[string]hclog.Logger),
	}
	l.Logger = hclog.New(&hclog.LoggerOptions{
		Output: l,
		// doesn't matter, we check verbosity levels on klog
		// before writting either way, so always Debug is fine
		Level: hclog.Trace,
		Name:  l.name,
	})
	return l
}
