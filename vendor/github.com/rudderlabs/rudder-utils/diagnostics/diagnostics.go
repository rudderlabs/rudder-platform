package diagnostics

import (
	"time"

	"github.com/rudderlabs/analytics-go"
	"github.com/rudderlabs/rudder-utils/utils/misc"
)

const (
	StartTime              = "diagnosis_start_time"
	InstanceId             = "server_instance_id"
	ServerStart            = "server_start"
	ConfigProcessed        = "config_processed"
	SourcesCount           = "no_of_sources"
	DesitanationCount      = "no_of_destinations"
	ServerStarted          = "server_started"
	ConfigIdentify         = "identify"
	GatewayEvents          = "gateway_events"
	GatewaySuccess         = "gateway_success"
	GatewayFailure         = "gateway_failure"
	RouterEvents           = "router_events"
	RouterAborted          = "router_aborted"
	RouterRetries          = "router_retries"
	RouterSuccess          = "router_success"
	RouterFailed           = "router_failed"
	RouterDestination      = "router_destination"
	RouterAttemptNum       = "router_attempt_num"
	RouterCompletedTime    = "router_average_job_time"
	BatchRouterEvents      = "batch_router_events"
	BatchRouterSuccess     = "batch_router_success"
	BatchRouterFailed      = "batch_router_failed"
	BatchRouterDestination = "batch_router_destination"
	UserID                 = "user_id"
	ErrorCode              = "error_code"
	ErrorResponse          = "error_response"
)

var (
	EnableDiagnostics               bool
	endpoint                        string
	writekey                        string
	EnableServerStartMetric         bool
	EnableConfigIdentifyMetric      bool
	EnableServerStartedMetric       bool
	EnableConfigProcessedMetric     bool
	EnableGatewayMetric             bool
	EnableRouterMetric              bool
	EnableBatchRouterMetric         bool
	EnableDestinationFailuresMetric bool
)
var Diagnostics DiagnosticsI

type DiagnosticsI interface {
	Track(event string, properties map[string]interface{})
	DisableMetrics(enableMetrics bool)
	Identify(properties map[string]interface{})
}
type diagnostics struct {
	Client     analytics.Client
	StartTime  time.Time
	UniqueId   string
	UserId     string
	InstanceId string
	config     ConfigDiagnostics
}

type ConfigDiagnostics struct {
	EnableDiagnostics               bool
	Endpoint                        string
	Writekey                        string
	EnableServerStartMetric         bool
	EnableConfigIdentifyMetric      bool
	EnableServerStartedMetric       bool
	EnableConfigProcessedMetric     bool
	EnableGatewayMetric             bool
	EnableRouterMetric              bool
	EnableBatchRouterMetric         bool
	EnableDestinationFailuresMetric bool
	InstanceID                      string
}

var DefaultConfigDiagnostics = ConfigDiagnostics{EnableDiagnostics: true, Endpoint: "https://rudderstack-dataplane.rudderstack.com", Writekey: "1aWPBIROQvFYW9FHxgc03nUsLza", EnableServerStartMetric: true, EnableConfigIdentifyMetric: true, EnableServerStartedMetric: true, EnableConfigProcessedMetric: true, EnableGatewayMetric: true, EnableRouterMetric: true, EnableBatchRouterMetric: true, EnableDestinationFailuresMetric: true, InstanceID: "1"}

// func init() {
// 	loadConfig()
// }

func LoadConfig(configList ...interface{}) {
	config := checkAndValidateConfig((configList))
	EnableDiagnostics = config.EnableDiagnostics
	endpoint = config.Endpoint
	writekey = config.Writekey
	EnableServerStartMetric = config.EnableServerStartMetric
	EnableConfigIdentifyMetric = config.EnableConfigIdentifyMetric
	EnableServerStartedMetric = config.EnableServerStartedMetric
	EnableConfigProcessedMetric = config.EnableConfigProcessedMetric
	EnableGatewayMetric = config.EnableGatewayMetric
	EnableRouterMetric = config.EnableRouterMetric
	EnableBatchRouterMetric = config.EnableBatchRouterMetric
	EnableDestinationFailuresMetric = config.EnableDestinationFailuresMetric
	Diagnostics = newDiagnostics(config)
}

// newDiagnostics return new instace of diagnostics

func checkAndValidateConfig(configList []interface{}) ConfigDiagnostics {
	if len(configList) != 1 {
		return DefaultConfigDiagnostics
	}
	switch configList[0].(type) {
	case ConfigDiagnostics:
		return configList[0].(ConfigDiagnostics)
	default:
		return DefaultConfigDiagnostics
	}
}

func newDiagnostics(config ConfigDiagnostics) *diagnostics {
	instanceId := config.InstanceID

	client := analytics.New(config.Writekey, config.Endpoint)
	return &diagnostics{
		config:     config,
		InstanceId: instanceId,
		Client:     client,
		StartTime:  time.Now(),
		UniqueId:   misc.GetMD5Hash(misc.GetMacAddress()),
	}
}

func (d *diagnostics) Track(event string, properties map[string]interface{}) {
	if d.config.EnableDiagnostics {
		properties[StartTime] = d.StartTime
		properties[InstanceId] = d.InstanceId

		d.Client.Enqueue(
			analytics.Track{
				Event:       event,
				Properties:  properties,
				AnonymousId: d.UniqueId,
				UserId:      d.UserId,
			},
		)
	}
}

// Deprecated! Use instance of diagnostics instead;
func Track(event string, properties map[string]interface{}) {
	Diagnostics.Track(event, properties)
}

func (d *diagnostics) DisableMetrics(enableMetrics bool) {
	if !enableMetrics {
		d.config.EnableServerStartedMetric = false
		d.config.EnableConfigProcessedMetric = false
		d.config.EnableGatewayMetric = false
		d.config.EnableRouterMetric = false
		d.config.EnableBatchRouterMetric = false
		d.config.EnableDestinationFailuresMetric = false
	}
}

// Deprecated! Use instance of diagnostics instead;
func DisableMetrics(enableMetrics bool) {
	Diagnostics.DisableMetrics(enableMetrics)
}

func (d *diagnostics) Identify(properties map[string]interface{}) {
	if d.config.EnableDiagnostics {
		// add in traits
		if val, ok := properties[ConfigIdentify]; ok {
			d.UserId = val.(string)
		}
		d.Client.Enqueue(
			analytics.Identify{
				AnonymousId: d.UniqueId,
				UserId:      d.UserId,
			},
		)
	}
}

// Deprecated! Use instance of diagnostics instead;
func Identify(properties map[string]interface{}) {
	Diagnostics.Identify(properties)
}
