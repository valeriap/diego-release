package loggregator_v2_test

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/loggregator_v2"
	"code.cloudfoundry.org/loggregator_v2/fakes"
	lfake "github.com/cloudfoundry/dropsonde/log_sender/fake"
	mfake "github.com/cloudfoundry/dropsonde/metric_sender/fake"
	"github.com/cloudfoundry/dropsonde/metrics"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/cloudfoundry/dropsonde/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testHandler struct{}

var _ = Describe("Client", func() {
	var (
		config    loggregator_v2.MetronConfig
		logger    *lagertest.TestLogger
		client    loggregator_v2.Client
		clientErr error
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("loggregator-client")
	})

	JustBeforeEach(func() {
		client, clientErr = loggregator_v2.NewClient(logger, config)
	})

	Context("when v2 api is disabled", func() {
		var (
			logSender    *lfake.FakeLogSender
			metricSender *mfake.FakeMetricSender
		)

		BeforeEach(func() {
			logSender = &lfake.FakeLogSender{}
			metricSender = &mfake.FakeMetricSender{}
			config.UseV2API = false
			logs.Initialize(logSender)
			metrics.Initialize(metricSender, nil)
		})

		It("sends app logs", func() {
			client.SendAppLog("app-id", "message", "source-type", "source-instance")
			Expect(logSender.GetLogs()).To(ConsistOf(lfake.Log{AppId: "app-id", Message: "message",
				SourceType: "source-type", SourceInstance: "source-instance", MessageType: "OUT"}))
		})

		It("sends app error logs", func() {
			client.SendAppErrorLog("app-id", "message", "source-type", "source-instance")
			Expect(logSender.GetLogs()).To(ConsistOf(lfake.Log{AppId: "app-id", Message: "message",
				SourceType: "source-type", SourceInstance: "source-instance", MessageType: "ERR"}))
		})

		It("sends app metrics", func() {
			metric := events.ContainerMetric{
				ApplicationId: proto.String("app-id"),
			}
			client.SendAppMetrics(&metric)
			Expect(metricSender.Events()).To(ConsistOf(&metric))
		})
	})

	Context("when v2 api is enabled", func() {
		var (
			server       *grpc.Server
			senderServer *fakes.FakeIngressServer
			receivers    chan loggregator_v2.Ingress_SenderServer
		)

		FContext("the cert or key path are invalid", func() {
			BeforeEach(func() {
				config.CertPath = "/some/invalid/path"
			})

			It("returns an error", func() {
				Expect(clientErr).To(HaveOccurred(), "client didn't return an error")
			})
		})

		FContext("the ca cert path is invalid", func() {
			BeforeEach(func() {
				config.CACertPath = "/some/invalid/path"
			})

			It("returns an error", func() {
				Expect(clientErr).To(HaveOccurred(), "client didn't return an error")
			})
		})

		FContext("the ca cert is invalid", func() {
			BeforeEach(func() {
				config.CACertPath = "fixtures/invalid-ca.crt"
			})

			It("returns an error", func() {
				Expect(clientErr).To(HaveOccurred(), "client didn't return an error")
			})
		})

		FContext("cannot connecto to the server", func() {
			BeforeEach(func() {
				config.APIPort = 1234
			})

			It("returns an error", func() {
				Expect(clientErr).To(HaveOccurred(), "client didn't return an error")
			})
		})

		BeforeEach(func() {
			cert, err := tls.LoadX509KeyPair("fixtures/metron.crt", "fixtures/metron.key")
			Expect(err).NotTo(HaveOccurred())
			tlsConfig := &tls.Config{
				Certificates:       []tls.Certificate{cert},
				ClientAuth:         tls.RequestClientCert,
				InsecureSkipVerify: false,
			}
			caCertBytes, err := ioutil.ReadFile("fixtures/CA.crt")
			Expect(err).NotTo(HaveOccurred())
			caCertPool := x509.NewCertPool()
			Expect(err).NotTo(HaveOccurred())
			caCertPool.AppendCertsFromPEM(caCertBytes)
			tlsConfig.RootCAs = caCertPool
			server = grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
			listener, err := net.Listen("tcp4", "localhost:0")
			Expect(err).NotTo(HaveOccurred())
			senderServer = &fakes.FakeIngressServer{}
			receivers = make(chan loggregator_v2.Ingress_SenderServer)
			senderServer.SenderStub = func(recv loggregator_v2.Ingress_SenderServer) error {
				receivers <- recv
				return nil
			}
			loggregator_v2.RegisterIngressServer(server, senderServer)
			go server.Serve(listener)
			port := listener.Addr().(*net.TCPAddr).Port
			Expect(err).NotTo(HaveOccurred())
			config = loggregator_v2.MetronConfig{
				UseV2API:   true,
				APIPort:    port,
				CACertPath: "fixtures/CA.crt",
				CertPath:   "fixtures/client.crt",
				KeyPath:    "fixtures/client.key",
			}
		})

		AfterEach(func() {
			server.GracefulStop()
		})

		It("sends app logs", func() {
			Consistently(func() error {
				return client.SendAppLog("app-id", "message", "source-type", "source-instance")
			}).Should(Succeed())
			var recv loggregator_v2.Ingress_SenderServer
			Eventually(receivers).Should(Receive(&recv))
			env, err := recv.Recv()
			Expect(err).NotTo(HaveOccurred())
			log := env.GetLog()
			Expect(log).NotTo(BeNil())
			Expect(log.GetPayload()).To(Equal([]byte("message")))
			Expect(log.Type).To(Equal(loggregator_v2.Log_OUT))
		})

		It("sends app error logs", func() {
			// TODO: why do we need this ?
			Consistently(func() error {
				return client.SendAppErrorLog("app-id", "message", "source-type", "source-instance")
			}).Should(Succeed())
			var recv loggregator_v2.Ingress_SenderServer
			Eventually(receivers).Should(Receive(&recv))
			env, err := recv.Recv()
			Expect(err).NotTo(HaveOccurred())
			log := env.GetLog()
			Expect(log).NotTo(BeNil())
			Expect(log.GetPayload()).To(Equal([]byte("message")))
			Expect(log.Type).To(Equal(loggregator_v2.Log_ERR))
		})

		It("sends app metrics", func() {
			metric := events.ContainerMetric{
				ApplicationId: proto.String("app-id"),
				CpuPercentage: proto.Float64(10.0),
			}
			Consistently(func() error {
				return client.SendAppMetrics(&metric)
			}).Should(Succeed())
			var recv loggregator_v2.Ingress_SenderServer
			Eventually(receivers).Should(Receive(&recv))
			env, err := recv.Recv()
			Expect(err).NotTo(HaveOccurred())
			metrics := env.GetGauge()
			Expect(metrics).NotTo(BeNil())
			Expect(metrics.GetMetrics()).NotTo(BeNil())
			Expect(env.GetSourceId()).To(Equal("app-id"))
			Expect(metrics.GetMetrics()["cpu"].GetValue()).To(Equal(10.0))
			Expect(metrics.GetMetrics()["cpu"].GetUnit()).To(Equal("nano")) // TODO: what should this be ?
		})
	})
})
