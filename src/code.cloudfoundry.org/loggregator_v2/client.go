package loggregator_v2

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"

	"code.cloudfoundry.org/lager"

	"github.com/cloudfoundry/sonde-go/events"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

//go:generate bash -c "protoc ../loggregator-api/v2/*.proto --go_out=plugins=grpc:. --proto_path=../loggregator-api/v2"
//go:generate counterfeiter -o fakes/fake_client.go . Client
type Client interface {
	SendAppLog(appID, message, sourceType, sourceInstance string) error
	SendAppErrorLog(appID, message, sourceType, sourceInstance string) error
	SendAppMetrics(metrics *events.ContainerMetric) error
}

type MetronConfig struct {
	UseV2API      bool   `json:"metron_use_v2_api"`
	APIPort       int    `json:"metron_api_port"`
	CACertPath    string `json:"metron_ca_path"`
	CertPath      string `json:"metron_cert_path"`
	KeyPath       string `json:"metron_key_path"`
	DropsondePort int    `json:"dropsonde_port"`
}

func NewClient(logger lager.Logger, config MetronConfig) (Client, error) {
	if !config.UseV2API {
		return &dropsondeClient{}, nil
	}
	logger.Info("creating-grpc-client", lager.Data{"config": config})
	address := fmt.Sprintf("localhost:%d", config.APIPort)
	cert, err := tls.LoadX509KeyPair(config.CertPath, config.KeyPath)
	if err != nil {
		logger.Error("cannot-load-certs", err)
		return nil, err
	}
	tlsConfig := &tls.Config{
		ServerName:         "metron",
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false,
	}
	caCertBytes, err := ioutil.ReadFile(config.CACertPath)
	if err != nil {
		logger.Error("failed-to-read-ca-cert", err)
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		logger.Error("failed-to-append-ca-cert", err)
		return nil, errors.New("cannot parse ca cert")
	}
	tlsConfig.RootCAs = caCertPool
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		logger.Error("failed-to-create-grpc-client", err)
		return nil, err
	}
	c := NewIngressClient(conn)
	sender, err := c.Sender(context.Background())
	if err != nil {
		logger.Error("failed-to-create-grpc-sender", err)
		return nil, err
	}
	return &grpcClient{sender}, nil
}
