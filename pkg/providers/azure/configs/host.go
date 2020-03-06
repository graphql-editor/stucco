package configs

// Aggregator in host.json
type Aggregator struct {
	BatchSize    int    `json:"batchSize,omitempty"`
	FlushTimeout string `json:"flushTimeout,omitempty"`
}

// HSTS for HTTPExtension
type HSTS struct {
	IsEnabled *bool  `json:"isEnabled,omitempty"`
	MaxAge    string `json:"maxAge,omitempty"`
}

// HTTPExtension in host.json extensions
type HTTPExtension struct {
	RoutePrefix             *string           `json:"routePrefix,omitempty"`
	MaxOutstandingRequest   int               `json:"maxOutstandingRequests,omitempty"`
	MaxConcurrentRequest    int               `json:"maxConcurrentRequests,omitempty"`
	DynamicThrottlesEnabled *bool             `json:"dynamicThrottlesEnabled,omitempty"`
	HSTS                    *HSTS             `json:"hsts,omitempty"`
	CustomHeaders           map[string]string `json:"customHeaders,omitempty"`
}

// Extensions in host.json
type Extensions struct {
	// TODO: defined all extensions
	CosmosDB    map[string]interface{} `json:"cosmosDB,omitempty"`
	DurableTask map[string]interface{} `json:"durableTask,omitempty"`
	EventHubs   map[string]interface{} `json:"eventHubs,omitempty"`
	HTTP        *HTTPExtension         `json:"http,omitempty"`
	Queues      map[string]interface{} `json:"queues,omitempty"`
	SendGrid    map[string]interface{} `json:"sendGrid,omitempty"`
	ServiceBus  map[string]interface{} `json:"serviceBus,omitempty"`
}

// ExtensionBundle in host.json
type ExtensionBundle struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// HealthMonitor in host.json
type HealthMonitor struct {
	Enabled             *bool   `json:"enabled,omitempty"`
	HealthCheckInterval string  `json:"healthCheckInterval,omitempty"`
	HealthCheckWindow   string  `json:"healthCheckWindow,omitempty"`
	HealthCheckTreshold int     `json:"healthCheckTreshold,omitempty"`
	CounterTreshold     float64 `json:"counterTreshold,omitempty"`
}

// LoggingMode in logging
type LoggingMode string

const (
	// NeverLoggingMode logging mode
	NeverLoggingMode LoggingMode = "never"
	// AlwaysLoggingMode logging mode
	AlwaysLoggingMode LoggingMode = "always"
	// DebugOnlyLoggingMode logging mode
	DebugOnlyLoggingMode LoggingMode = "debugOnly"
)

// LogLevel in logging
type LogLevel string

const (
	// TraceLogLevel log level
	TraceLogLevel LogLevel = "Trace"
	// DebugLogLevel log level
	DebugLogLevel LogLevel = "Debug"
	// InformationLogLevel log level
	InformationLogLevel LogLevel = "Information"
	// WarningLogLevel log level
	WarningLogLevel LogLevel = "Warning"
	// ErrorLogLevel log level
	ErrorLogLevel LogLevel = "Error"
	// CriticalLogLevel log level
	CriticalLogLevel LogLevel = "Critical"
	// NoneLogLevel log level
	NoneLogLevel LogLevel = "None"
)

// Logging in host.json
type Logging struct {
	// TODO: application insights
	FileLoggingMode LoggingMode         `json:"fileLoggingMode,omitempty"`
	LogLevel        map[string]LogLevel `json:"logLevel,omitempty"`
}

// ManagedDependency in host.json
type ManagedDependency struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// Singleton in host.json
type Singleton struct {
	LockPeriod                          string `json:"lockPeriod,omitempty"`
	ListenerLockPeriod                  string `json:"listenerLockPeriod,omitempty"`
	ListenerLockRecoveryPollingInterval string `json:"listenerLockPollingInterval,omitempty"`
	LockAcquisitionTimeout              string `json:"lockAcquisitionTimeout,omitempty"`
	LockAcquisitionPollingInterval      string `json:"lockAcquisitionPollingInterval,omitempty"`
}

// Host is host.json
type Host struct {
	Version           string             `json:"version"`
	Aggregator        *Aggregator        `json:"aggregator,omitempty"`
	Extensions        *Extensions        `json:"extensions,omitempty"`
	ExtensionBundle   *ExtensionBundle   `json:"extensionBundle,omitempty"`
	Functions         []string           `json:"functions,omitempty"`
	FunctionTimeout   string             `json:"functionTimeout,omitempty"`
	Logging           *Logging           `json:"logging,omitempty"`
	ManagedDependency *ManagedDependency `json:"managedDependency,omitempty"`
	Singleton         *Singleton         `json:"singleton,omitempty"`
	WatchDirectories  []string           `json:"watchDirectories,omitempty"`
}
