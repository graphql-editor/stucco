package plugin

import (
	"bytes"
	"sync"

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

func (k *klogLogger) clean(lv hclog.Level) string {
	bs := k.Bytes()
	if len(bs) < k.trimTimeFormat+_trimLevel[lv]+2 {
		// badly formatted log
		// return it without doing anything
		return string(bs)
	}
	return string(bs)[k.trimTimeFormat+_trimLevel[lv]+2:]
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
