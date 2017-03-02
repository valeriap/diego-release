package loggregator_v2

import (
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

//go:generate counterfeiter -o fakes/fake_ingress_server.go . IngressServer
//go:generate counterfeiter -o fakes/fake_ingress_sender_server.go . Ingress_SenderServer

type grpcClient struct {
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
	c.client.Send(&Envelope{
		Timestamp: int64(time.Now().Nanosecond()),
		SourceId:  m.GetApplicationId(),
		Message: &Envelope_Gauge{
			Gauge: &Gauge{
				Metrics: map[string]*GaugeValue{
					"cpu": &GaugeValue{
						Unit:  "nano",
						Value: m.GetCpuPercentage(),
					},
				},
			},
		},
	})
	return nil
}
