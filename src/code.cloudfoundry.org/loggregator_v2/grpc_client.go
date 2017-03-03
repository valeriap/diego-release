package loggregator_v2

import (
	"time"

	"code.cloudfoundry.org/lager"

	"github.com/cloudfoundry/sonde-go/events"
)

//go:generate counterfeiter -o fakes/fake_ingress_server.go . IngressServer
//go:generate counterfeiter -o fakes/fake_ingress_sender_server.go . Ingress_SenderServer

type grpcClient struct {
	logger lager.Logger
	client Ingress_SenderClient
}

func createLogEnvelope(appID, message, sourceType, sourceInstance string, logType Log_Type) *Envelope {
	return &Envelope{
		Timestamp: int64(time.Now().Nanosecond()),
		SourceId:  appID,
		Message: &Envelope_Log{
			Log: &Log{
				Payload: []byte(message),
				Type:    logType,
			},
		},
		Tags: map[string]*Value{
			"source_type": &Value{
				Data: &Value_Text{
					Text: sourceType,
				},
			},
			"source_instance": &Value{
				Data: &Value_Text{
					Text: sourceInstance,
				},
			},
		},
	}
}

func (c *grpcClient) SendAppLog(appID, message, sourceType, sourceInstance string) error {
	return c.client.Send(createLogEnvelope(appID, message, sourceType, sourceInstance, Log_OUT))
}

func (c *grpcClient) SendAppErrorLog(appID, message, sourceType, sourceInstance string) error {
	return c.client.Send(createLogEnvelope(appID, message, sourceType, sourceInstance, Log_ERR))
}

func (c *grpcClient) SendAppMetrics(m *events.ContainerMetric) error {
	c.logger.Info("grpc-logger-send-metric", lager.Data{"app-id": m.GetApplicationId()})
	c.client.Send(&Envelope{
		Timestamp: int64(time.Now().Nanosecond()),
		SourceId:  m.GetApplicationId(),
		Message: &Envelope_Gauge{
			Gauge: &Gauge{
				Metrics: map[string]*GaugeValue{
					"instance_index": &GaugeValue{
						Unit:  "index",
						Value: float64(m.GetInstanceIndex()),
					},
					"cpu": &GaugeValue{
						Unit:  "nano",
						Value: float64(m.GetCpuPercentage()),
					},
					"memory": &GaugeValue{
						Unit:  "bytes",
						Value: float64(m.GetMemoryBytes()),
					},
					"disk": &GaugeValue{
						Unit:  "bytes",
						Value: float64(m.GetDiskBytes()),
					},
					"memory_quota": &GaugeValue{
						Unit:  "bytes",
						Value: float64(m.GetMemoryBytesQuota()),
					},
					"disk_quota": &GaugeValue{
						Unit:  "bytes",
						Value: float64(m.GetDiskBytesQuota()),
					},
				},
			},
		},
	})
	return nil
}
