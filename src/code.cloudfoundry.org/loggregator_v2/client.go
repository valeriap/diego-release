package loggregator_v2

import "github.com/cloudfoundry/sonde-go/events"

//go:generate counterfeiter -o fakes/fake_client.go . Client
type Client interface {
	SendAppLog(appID, message, sourceType, sourceInstance string) error
	SendAppErrorLog(appID, message, sourceType, sourceInstance string) error
	SendAppMetrics(metrics *events.ContainerMetric) error
}

type MetronConfig struct {
	UseV2API      bool   `json:"use_v2_api"`
	V2APIPort     int    `json:"v2_api_port"`
	CACertPath    string `json:"ca_cert_path"`
	CertPath      string `json:"cert_path"`
	KeyPath       string `json:"key_path"`
	DropsondePort int    `json:"dropsonde_port"`
}

func NewClient(config MetronConfig) Client {
	if !config.UseV2API {
		return &dropsondeClient{}
	}
	panic("undefined")
}
