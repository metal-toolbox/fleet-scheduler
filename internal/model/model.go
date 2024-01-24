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
	ServerserviceAPI APIKind = "serverservice"
	ConditionsAPI    APIKind = "conditions"

	StoreKindServerservice StoreKind = "serverservice"

	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
	LogLevelTrace LogLevel = "trace"

	ConcurrencyDefault = 5
	ProfilingEndpoint  = "localhost:9091"
	MetricsEndpoint    = "0.0.0.0:9090"
)
