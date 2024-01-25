package model

type (
	AppKind   string
	StoreKind string

	InventoryMethod string

	// LogLevel is the logging level string.
	LogLevel string

	CollectorError string
)

type (
	APIKind string
)

const (
	AppName string = "fleetscheduler"

	DefaultFleetDBClientID      = "fleetscheduler-serverservice-api"
	DefaultConditionOrcClientID = "fleetscheduler-condition-api"

	ServerserviceAPI APIKind = "serverservice"
	ConditionsAPI    APIKind = "conditions"

	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
	LogLevelTrace LogLevel = "trace"

	ConcurrencyDefault = 5
)
