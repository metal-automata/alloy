package model

import (
	"net"

	"github.com/bmc-toolbox/common"

	cptypes "github.com/metal-toolbox/conditionorc/pkg/types"
)

type (
	AppKind   string
	StoreKind string

	// LogLevel is the logging level string.
	LogLevel string
)

const (
	AppName                  = "alloy"
	AppKindInband    AppKind = "inband"
	AppKindOutOfBand AppKind = "outofband"

	// conditions reconciled by this controller
	InventoryOutofband cptypes.ConditionKind = "inventoryOutofband"

	StoreKindCsv           StoreKind = "csv"
	StoreKindServerservice StoreKind = "serverservice"
	StoreKindMock          StoreKind = "mock"

	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
	LogLevelTrace LogLevel = "trace"

	ConcurrencyDefault = 5
	ProfilingEndpoint  = "localhost:9091"
	MetricsEndpoint    = "0.0.0.0:9090"
	// EnvVarDumpFixtures when enabled, will dump data for assets, to be used as fixture data.
	EnvVarDumpFixtures = "DEBUG_DUMP_FIXTURES"
	// EnvVarDumpDiffers when enabled, will dump component differ data for debugging
	// differences identified in component objects in the publish package.
	EnvVarDumpDiffers = "DEBUG_DUMP_DIFFERS"
)

// Asset represents attributes of an asset retrieved from the asset store
type Asset struct {
	// Inventory collected from the device
	Inventory *common.Device
	// The device metadata attribute
	Metadata map[string]string
	// BIOS configuration
	BiosConfig map[string]string
	// The device ID from the inventory store
	ID string
	// The device vendor attribute
	Vendor string
	// The device model attribute
	Model string
	// The device serial attribute
	Serial string
	// The datacenter facility attribute from the configuration
	Facility string
	// Username is the BMC login username from the inventory store
	BMCUsername string
	// Password is the BMC login password from the inventory store
	BMCPassword string
	// Errors is a map of errors,
	// where the key is the stage at which the error occurred,
	// and the value is the error.
	Errors map[string]string
	// Address is the BMC IP address from the inventory store
	BMCAddress net.IP
}

// IncludeError includes the given error key and value in the asset
// which is then available to the publisher for reporting.
func (a *Asset) IncludeError(key, value string) {
	a.Errors[key] = value
}
