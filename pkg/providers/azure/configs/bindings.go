package configs

// Direction enum
type Direction string

const (
	// InDirection direction
	InDirection Direction = "in"
	// OutDirection direction
	OutDirection Direction = "out"
	// InOutDirection direction
	InOutDirection Direction = "inout"
)

// DataType enum
type DataType string

const (
	// StringDataType data type
	StringDataType DataType = "string"
	// BinaryDataType data type
	BinaryDataType DataType = "binary"
	// StreamDataType data type
	StreamDataType DataType = "stream"
)

// BindingType enum contains all predefined binding types
type BindingType string

// Dynamic returns true if binding is dynamic binding
func (b BindingType) Dynamic() bool {
	switch b {
	case ServiceBusTrigger, ServiceBus, BlobTrigger, Blob, ManualTrigger, EventHubTrigger, EventHub, TimerTrigger, QueueTrigger, Queue, HTTPTrigger, HTTP, MobileTable, DocumentDB, Table, NotificationHub, TwilioSms, SendGrid:
		return false
	default:
		return true
	}
}

const (
	// ServiceBusTrigger binding type
	ServiceBusTrigger BindingType = "serviceBusTrigger"
	// ServiceBus binding type
	ServiceBus BindingType = "serviceBus"
	// BlobTrigger binding type
	BlobTrigger BindingType = "blobTrigger"
	// Blob binding type
	Blob BindingType = "blob"
	// ManualTrigger binding type
	ManualTrigger BindingType = "manualTrigger"
	// EventHubTrigger binding type
	EventHubTrigger BindingType = "eventHubTrigger"
	// EventHub binding type
	EventHub BindingType = "eventHub"
	// TimerTrigger binding type
	TimerTrigger BindingType = "timerTrigger"
	// QueueTrigger binding type
	QueueTrigger BindingType = "queueTrigger"
	// Queue binding type
	Queue BindingType = "queue"
	// HTTPTrigger binding type
	HTTPTrigger BindingType = "httpTrigger"
	// HTTP binding type
	HTTP BindingType = "http"
	// MobileTable binding type
	MobileTable BindingType = "mobileTable"
	// DocumentDB binding type
	DocumentDB BindingType = "documentDB"
	// Table binding type
	Table BindingType = "table"
	// NotificationHub binding type
	NotificationHub BindingType = "notificationHub"
	// TwilioSms binding type
	TwilioSms BindingType = "twilioSms"
	// SendGrid binding type
	SendGrid BindingType = "sendGrid"
)

// AccessRights enum
type AccessRights string

const (
	// ManageAccessRights access rights
	ManageAccessRights AccessRights = "manage"
	// ListenAccessRights access rights
	ListenAccessRights AccessRights = "listen"
)

// AuthLevel enum
type AuthLevel string

const (
	// AnonymousAuthLevel auth level
	AnonymousAuthLevel AuthLevel = "anonymous"
	// FunctionAuthLevel auth level
	FunctionAuthLevel AuthLevel = "function"
	// AdminAuthLevel auth level
	AdminAuthLevel AuthLevel = "admin"
)

// Cardinality enum
type Cardinality string

const (
	// OneCardinality cardinality
	OneCardinality Cardinality = "one"
	// ManyCardinality cardinality
	ManyCardinality Cardinality = "many"
)

// Method enum
type Method string

const (
	// GetMethod method
	GetMethod Method = "get"
	// PostMethod method
	PostMethod Method = "post"
	// DeleteMethod method
	DeleteMethod Method = "delete"
	// HeadMethod method
	HeadMethod Method = "head"
	// PatchMethod method
	PatchMethod Method = "patch"
	// PutMethod method
	PutMethod Method = "put"
	// OptionsMethod method
	OptionsMethod Method = "options"
	// TraceMethod method
	TraceMethod Method = "trace"
)

// Platform enum
type Platform string

const (
	// ApnsPlatform platform
	ApnsPlatform Platform = "apns"
	// AdmPlatform platform
	AdmPlatform Platform = "adm"
	// GcmPlatform platform
	GcmPlatform Platform = "gcm"
	// WnsPlatform platform
	WnsPlatform Platform = "wns"
	// MpnsPlatform platform
	MpnsPlatform Platform = "mpns"
)

// Binding represents item in bindings array
type Binding struct {
	Name      string      `json:"name"`
	Type      BindingType `json:"type"`
	Direction Direction   `json:"direction"`
	DataType  DataType    `json:"dataType,omitempty"`

	AccessRights      AccessRights `json:"accessRights,omitempty"`
	AccountSid        string       `json:"accountSid,omitempty"`
	APIKey            string       `json:"apiKey,omitempty"`
	AuthLevel         AuthLevel    `json:"authLevel,omitempty"`
	AuthToken         string       `json:"authToken,omitempty"`
	Body              string       `json:"body,omitempty"`
	Cardinality       Cardinality  `json:"cardinality,omitempty"`
	CollectionName    string       `json:"collectionName,omitempty"`
	Connection        string       `json:"connection,omitempty"`
	ConsumerGroup     string       `json:"consumerGroup,omitempty"`
	CreateIfNotExists *bool        `json:"createIfNotExists,omitempty"`
	DatabaseName      string       `json:"databaseName,omitempty"`
	Filter            string       `json:"filter,omitempty"`
	From              string       `json:"from,omitempty"`
	HubName           string       `json:"hubName,omitempty"`
	ID                string       `json:"id,omitempty"`
	Methods           []Method     `json:"methods,omitempty"`
	PartitionKey      string       `json:"partitionKey,omitempty"`
	Path              string       `json:"path,omitempty"`
	Platform          Platform     `json:"platform,omitempty"`
	QueueName         string       `json:"queueName,omitempty"`
	Route             string       `json:"route,omitempty"`
	RowKey            string       `json:"rowKey,omitempty"`
	RunOnStartup      *bool        `json:"runOnStartup,omitempty"`
	Schedule          string       `json:"schedule,omitempty"`
	SQLQuery          string       `json:"sqlQuery,omitempty"`
	Subject           string       `json:"subject,omitempty"`
	SubscriptionName  string       `json:"subscriptionName,omitempty"`
	TableName         string       `json:"tableName,omitempty"`
	TagExpresion      string       `json:"tagExpression,omitempty"`
	Take              string       `json:"take,omitempty"`
	Text              string       `json:"text,omitempty"`
	To                string       `json:"to,omitempty"`
	TopicName         string       `json:"topicName,omitempty"`
	UseMonitor        *bool        `json:"useMonitor,omitempty"`
	WebHookType       string       `json:"webHookType,omitempty"`
}
