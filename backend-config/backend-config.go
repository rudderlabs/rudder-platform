package backendconfig

//go:generate mockgen -destination=../../mocks/config/backend-config/mock_backendconfig.go -package=mock_backendconfig github.com/rudderlabs/rudder-server/config/backend-config BackendConfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/rudderlabs/rudder-utils/diagnostics"
	"github.com/rudderlabs/rudder-utils/stats"

	"github.com/rudderlabs/rudder-utils/logger"
	"github.com/rudderlabs/rudder-utils/utils/types"

	"github.com/rudderlabs/rudder-utils/rruntime"
	"github.com/rudderlabs/rudder-utils/utils"
	"github.com/rudderlabs/rudder-utils/utils/sysUtils"
)

var (
	backendConfig                         BackendConfig
	isMultiWorkspace                      bool
	multiWorkspaceSecret                  string
	configBackendURL, workspaceToken      string
	pollInterval, regulationsPollInterval time.Duration
	configFromFile                        bool
	configJSONPath                        string
	curSourceJSON                         ConfigT
	curSourceJSONLock                     sync.RWMutex
	curRegulationJSON                     RegulationsT
	curRegulationJSONLock                 sync.RWMutex
	initializedLock                       sync.RWMutex
	initialized                           bool
	waitForRegulations                    bool
	LastSync                              string
	LastRegulationSync                    string
	maxRegulationsPerRequest              int
	configEnvReplacementEnabled           bool
	errorFilePath                         string
	configEnvHandler                      types.ConfigEnvI

	//DefaultBackendConfig will be initialized be Setup to either a WorkspaceConfig or MultiWorkspaceConfig.
	DefaultBackendConfig BackendConfig
	Http                 sysUtils.HttpI           = sysUtils.NewHttp()
	pkgLogger            logger.LoggerI           = logger.NewLogger(logger.DefaultConfigLogger).Child("backend-config")
	IoUtil               sysUtils.IoUtilI         = sysUtils.NewIoUtil()
	Diagnostics          diagnostics.DiagnosticsI = diagnostics.Diagnostics
)

var Eb utils.PublishSubscriber = new(utils.EventBus)

// Topic refers to a subset of backend config's updates, received after subscribing using the backend config's Subscribe function.
type Topic string

type Regulation string

const (
	/*TopicBackendConfig topic provides updates on full backend config, via Subscribe function */
	TopicBackendConfig Topic = "backendConfig"

	/*TopicProcessConfig topic provides updates on backend config of processor enabled destinations, via Subscribe function */
	TopicProcessConfig Topic = "processConfig"

	/*TopicRegulations topic provides updates on regulations, via Subscribe function */
	TopicRegulations Topic = "regulations"

	/*RegulationSuppress refers to Suppress Regulation */
	RegulationSuppress Regulation = "Suppress"

	//TODO Will add support soon.
	/*RegulationDelete refers to Suppress and Delete Regulation */
	RegulationDelete Regulation = "Delete"

	/*RegulationSuppressAndDelete refers to Suppress and Delete Regulation */
	RegulationSuppressAndDelete Regulation = "Suppress_With_Delete"
)

type DestinationDefinitionT struct {
	ID            string
	Name          string
	DisplayName   string
	Config        map[string]interface{}
	ResponseRules map[string]interface{}
}

type SourceDefinitionT struct {
	ID       string
	Name     string
	Category string
}

type DestinationT struct {
	ID                    string
	Name                  string
	DestinationDefinition DestinationDefinitionT
	Config                map[string]interface{}
	Enabled               bool
	Transformations       []TransformationT
	IsProcessorEnabled    bool
}

type SourceT struct {
	ID               string
	Name             string
	SourceDefinition SourceDefinitionT
	Config           map[string]interface{}
	Enabled          bool
	WorkspaceID      string
	Destinations     []DestinationT
	WriteKey         string
}

type WorkspaceRegulationT struct {
	ID             string
	RegulationType string
	WorkspaceID    string
	UserID         string
}

type SourceRegulationT struct {
	ID             string
	RegulationType string
	WorkspaceID    string
	SourceID       string
	UserID         string
}

type ConfigT struct {
	EnableMetrics bool       `json:"enableMetrics"`
	WorkspaceID   string     `json:"workspaceId"`
	Sources       []SourceT  `json:"sources"`
	Libraries     LibrariesT `json:"libraries"`
}

type RegulationsT struct {
	WorkspaceRegulations []WorkspaceRegulationT `json:"workspaceRegulations"`
	SourceRegulations    []SourceRegulationT    `json:"sourceRegulations"`
}

type WRegulationsT struct {
	WorkspaceRegulations []WorkspaceRegulationT `json:"workspaceRegulations"`
	Start                int                    `json:"start"`
	Limit                int                    `json:"limit"`
	Size                 int                    `json:"size"`
	End                  bool                   `json:"end"`
	Next                 int                    `json:"next"`
}

type SRegulationsT struct {
	SourceRegulations []SourceRegulationT `json:"sourceRegulations"`
	Start             int                 `json:"start"`
	Limit             int                 `json:"limit"`
	Size              int                 `json:"size"`
	End               bool                `json:"end"`
	Next              int                 `json:"next"`
}

type TransformationT struct {
	VersionID string
}

type LibraryT struct {
	VersionID string
}

type LibrariesT []LibraryT

type BackendConfig interface {
	SetUp()
	Get() (ConfigT, bool)
	GetRegulations() (RegulationsT, bool)
	GetWorkspaceIDForWriteKey(string) string
	GetWorkspaceLibrariesForWorkspaceID(string) LibrariesT
	WaitForConfig()
	Subscribe(channel chan utils.DataEvent, topic Topic)
}
type CommonBackendConfig struct {
	configEnvHandler types.ConfigEnvI
}

type BackendConfigSetup struct {
	IsMultiWorkspace            bool
	MultiWorkspaceSecret        string
	ConfigBackendUrl            string
	WorkSpaceToken              string
	PollInterval                time.Duration
	RegulationsPollInterval     time.Duration
	ConfigJSONPath              string
	ConfigFromFile              bool
	MaxRegulationsPerRequest    int
	ConfigEnvReplacementEnabled bool
	ErrorFilePath               string
	ConfigLogger                logger.ConfigLogger
	ConfigStats                 stats.ConfigStats
	ConfigDiagnostics           diagnostics.ConfigDiagnostics
}

var DefaultBackendConfigSetup = BackendConfigSetup{IsMultiWorkspace: false, MultiWorkspaceSecret: "password", ConfigBackendUrl: "https://api.rudderlabs.com", WorkSpaceToken: "", RegulationsPollInterval: 300 * time.Second, ConfigJSONPath: "/etc/rudderstack/workspaceConfig.json", ConfigFromFile: false, MaxRegulationsPerRequest: 1000, ConfigEnvReplacementEnabled: true, ErrorFilePath: "/tmp/error_store.json", ConfigLogger: logger.DefaultConfigLogger, ConfigStats: stats.DefaultConfigStats, ConfigDiagnostics: diagnostics.DefaultConfigDiagnostics}

func checkAndValidateConfig(configList []interface{}) BackendConfigSetup {
	if len(configList) != 1 {
		return DefaultBackendConfigSetup
	}
	switch configList[0].(type) {
	case BackendConfigSetup:
		return configList[0].(BackendConfigSetup)
	default:
		return DefaultBackendConfigSetup
	}
}

func loadConfig(configList ...interface{}) {

	config := checkAndValidateConfig(configList)
	// Rudder supporting multiple workspaces. false by default
	isMultiWorkspace = config.IsMultiWorkspace
	// Secret to be sent in basic auth for supporting multiple workspaces. password by default
	multiWorkspaceSecret = config.MultiWorkspaceSecret

	configBackendURL = config.ConfigBackendUrl
	workspaceToken = config.WorkSpaceToken

	pollInterval = config.PollInterval
	regulationsPollInterval = config.RegulationsPollInterval
	configJSONPath = config.ConfigJSONPath
	configFromFile = config.ConfigFromFile
	maxRegulationsPerRequest = config.MaxRegulationsPerRequest
	configEnvReplacementEnabled = config.ConfigEnvReplacementEnabled
	diagnostics.LoadConfig(config.ConfigDiagnostics)
	pkgLogger = logger.NewLogger(config.ConfigLogger).Child("backend-config")
	stats.Setup(config.ConfigStats)
	errorFilePath = config.ErrorFilePath
}

func MakePostRequest(url string, endpoint string, data interface{}) (response []byte, ok bool) {
	client := &http.Client{}
	backendURL := fmt.Sprintf("%s%s", url, endpoint)
	dataJSON, _ := json.Marshal(data)
	request, err := Http.NewRequest("POST", backendURL, bytes.NewBuffer(dataJSON))
	if err != nil {
		pkgLogger.Errorf("ConfigBackend: Failed to make request: %s, Error: %s", backendURL, err.Error())
		return []byte{}, false
	}

	request.SetBasicAuth(workspaceToken, "")
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)
	// Not handling errors when sending alert to victorops
	if err != nil {
		pkgLogger.Errorf("ConfigBackend: Failed to execute request: %s, Error: %s", backendURL, err.Error())
		return []byte{}, false
	}
	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		pkgLogger.Errorf("ConfigBackend: Got error response %d", resp.StatusCode)
	}

	body, err := IoUtil.ReadAll(resp.Body)
	defer resp.Body.Close()

	pkgLogger.Debugf("ConfigBackend: Successful %s", string(body))
	return body, true
}

func MakeBackendPostRequest(endpoint string, data interface{}) (response []byte, ok bool) {
	return MakePostRequest(configBackendURL, endpoint, data)
}

func init() {
	loadConfig()
}

func trackConfig(preConfig ConfigT, curConfig ConfigT) {
	Diagnostics.DisableMetrics(curConfig.EnableMetrics)
	if diagnostics.EnableConfigIdentifyMetric {
		if len(preConfig.Sources) == 0 && len(curConfig.Sources) > 0 {
			Diagnostics.Identify(map[string]interface{}{
				diagnostics.ConfigIdentify: curConfig.Sources[0].WorkspaceID,
			})
		}
	}
	if diagnostics.EnableConfigProcessedMetric {
		noOfSources := len(curConfig.Sources)
		noOfDestinations := 0
		for _, source := range curConfig.Sources {
			noOfDestinations = noOfDestinations + len(source.Destinations)
		}
		Diagnostics.Track(diagnostics.ConfigProcessed, map[string]interface{}{
			diagnostics.SourcesCount:      noOfSources,
			diagnostics.DesitanationCount: noOfDestinations,
		})
	}
}

func filterProcessorEnabledDestinations(config ConfigT) ConfigT {
	var modifiedConfig ConfigT
	modifiedConfig.Libraries = config.Libraries
	modifiedConfig.Sources = make([]SourceT, 0)
	for _, source := range config.Sources {
		destinations := make([]DestinationT, 0)
		for _, destination := range source.Destinations {
			pkgLogger.Debug(destination.Name, " IsProcessorEnabled: ", destination.IsProcessorEnabled)
			if destination.IsProcessorEnabled {
				destinations = append(destinations, destination)
			}
		}
		source.Destinations = destinations
		modifiedConfig.Sources = append(modifiedConfig.Sources, source)
	}
	return modifiedConfig
}

func regulationsUpdate(statConfigBackendError stats.RudderStats) {

	regulationJSON, ok := backendConfig.GetRegulations()
	if !ok {
		statConfigBackendError.Increment()
	}

	//sorting the regulationJSON.
	//json unmarshal does not guarantee order. For DeepEqual to work as expected, sorting is necessary
	sort.Slice(regulationJSON.WorkspaceRegulations[:], func(i, j int) bool {
		return regulationJSON.WorkspaceRegulations[i].ID < regulationJSON.WorkspaceRegulations[j].ID
	})
	sort.Slice(regulationJSON.SourceRegulations[:], func(i, j int) bool {
		return regulationJSON.SourceRegulations[i].ID < regulationJSON.SourceRegulations[j].ID
	})

	if ok && !reflect.DeepEqual(curRegulationJSON, regulationJSON) {
		pkgLogger.Info("Regulations changed")
		curRegulationJSONLock.Lock()
		curRegulationJSON = regulationJSON
		curRegulationJSONLock.Unlock()
		initializedLock.Lock() //Using initializedLock for waitForRegulations too.
		defer initializedLock.Unlock()
		waitForRegulations = false
		LastRegulationSync = time.Now().Format(time.RFC3339)
		Eb.Publish(string(TopicRegulations), regulationJSON)
	}
}

func configUpdate(statConfigBackendError stats.RudderStats) {

	sourceJSON, ok := backendConfig.Get()
	if !ok {
		statConfigBackendError.Increment()
	}

	//sorting the sourceJSON.
	//json unmarshal does not guarantee order. For DeepEqual to work as expected, sorting is necessary
	sort.Slice(sourceJSON.Sources[:], func(i, j int) bool {
		return sourceJSON.Sources[i].ID < sourceJSON.Sources[j].ID
	})

	if ok && !reflect.DeepEqual(curSourceJSON, sourceJSON) {
		pkgLogger.Info("Workspace Config changed")
		curSourceJSONLock.Lock()
		trackConfig(curSourceJSON, sourceJSON)
		filteredSourcesJSON := filterProcessorEnabledDestinations(sourceJSON)
		curSourceJSON = sourceJSON
		curSourceJSONLock.Unlock()
		initializedLock.Lock()
		defer initializedLock.Unlock()
		initialized = true
		LastSync = time.Now().Format(time.RFC3339)
		Eb.Publish(string(TopicProcessConfig), filteredSourcesJSON)
		Eb.Publish(string(TopicBackendConfig), sourceJSON)
	}
}

func pollConfigUpdate() {
	statConfigBackendError := stats.NewStat("config_backend.errors", stats.CountType)
	for {
		configUpdate(statConfigBackendError)
		time.Sleep(time.Duration(pollInterval))
	}
}

func pollRegulations() {
	statConfigBackendError := stats.NewStat("config_backend.errors", stats.CountType)
	for {
		regulationsUpdate(statConfigBackendError)
		time.Sleep(time.Duration(regulationsPollInterval))
	}
}

func GetConfig() ConfigT {
	return curSourceJSON
}

func GetWorkspaceIDForWriteKey(writeKey string) string {
	return backendConfig.GetWorkspaceIDForWriteKey(writeKey)
}

func GetWorkspaceLibrariesForWorkspaceID(workspaceId string) LibrariesT {
	return backendConfig.GetWorkspaceLibrariesForWorkspaceID(workspaceId)
}

/*
Subscribe subscribes a channel to a specific topic of backend config updates.
Deprecated: Use an instance of BackendConfig instead of static function
*/
func Subscribe(channel chan utils.DataEvent, topic Topic) {
	backendConfig.Subscribe(channel, topic)
}

/*
Subscribe subscribes a channel to a specific topic of backend config updates.
Channel will receive a new utils.DataEvent each time the backend configuration is updated.
Data of the DataEvent should be a backendconfig.ConfigT struct.
Available topics are:
- TopicBackendConfig: Will receive complete backend configuration
- TopicProcessConfig: Will receive only backend configuration of processor enabled destinations
- TopicRegulations: Will receeive all regulations
*/
func (bc *CommonBackendConfig) Subscribe(channel chan utils.DataEvent, topic Topic) {
	Eb.Subscribe(string(topic), channel)
	curSourceJSONLock.RLock()

	if topic == TopicProcessConfig {
		filteredSourcesJSON := filterProcessorEnabledDestinations(curSourceJSON)
		Eb.PublishToChannel(channel, string(topic), filteredSourcesJSON)
	} else if topic == TopicBackendConfig {
		Eb.PublishToChannel(channel, string(topic), curSourceJSON)
	} else if topic == TopicRegulations {
		Eb.PublishToChannel(channel, string(topic), curRegulationJSON)
	}
	curSourceJSONLock.RUnlock()
}

/*
WaitForConfig waits until backend config has been initialized
Deprecated: Use an instance of BackendConfig instead of static function
*/
func WaitForConfig() {
	backendConfig.WaitForConfig()
}

/*
WaitForConfig waits until backend config has been initialized
*/
func (bc *CommonBackendConfig) WaitForConfig() {
	for {
		initializedLock.RLock()
		if initialized && !waitForRegulations {
			initializedLock.RUnlock()
			break
		}
		initializedLock.RUnlock()
		pkgLogger.Info("Waiting for initializing backend config")
		time.Sleep(time.Duration(pollInterval))
	}
}

// Setup backend config

//Setup ... LoadConfig and Setup or Call Setup and initialise LoadConfig in this
func Setup(pollRegulations bool, configEnvHandler types.ConfigEnvI, configList ...interface{}) {
	loadConfig(configList)

	if isMultiWorkspace {
		backendConfig = new(MultiWorkspaceConfig)
	} else {
		backendConfig = new(WorkspaceConfig)
		backendConfig.(*WorkspaceConfig).CommonBackendConfig.configEnvHandler = configEnvHandler
	}

	backendConfig.SetUp()

	DefaultBackendConfig = backendConfig

	rruntime.Go(func() {
		pollConfigUpdate()
	}, errorFilePath)

	if pollRegulations {
		startRegulationPolling()
	}

	//admin.RegisterAdminHandler("BackendConfig", &BackendConfigAdmin{})
}

// startRegulationPolling - starts enterprise backend regulations polling
func startRegulationPolling() {
	waitForRegulations = true

	rruntime.Go(func() {
		pollRegulations()
	}, errorFilePath)
}
