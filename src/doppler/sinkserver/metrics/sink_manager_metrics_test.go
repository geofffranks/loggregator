package metrics_test

import (
	"doppler/sinks"
	"doppler/sinks/containermetric"
	"doppler/sinks/dump"
	"doppler/sinks/syslog"
	"doppler/sinks/websocket"
	"doppler/sinkserver/metrics"
	"time"

	"github.com/cloudfoundry/dropsonde/emitter/fake"
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	dropsondeMetrics "github.com/cloudfoundry/dropsonde/metrics"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SinkManagerMetrics", func() {
	var (
		fakeEventEmitter   = fake.NewFakeEventEmitter("doppler")
		sinkManagerMetrics *metrics.SinkManagerMetrics
		sink               sinks.Sink
	)

	BeforeSuite(func() {
		sender := metric_sender.NewMetricSender(fakeEventEmitter)
		batcher := metricbatcher.New(sender, 100*time.Millisecond)
		dropsondeMetrics.Initialize(sender, batcher)
	})

	BeforeEach(func() {
		fakeEventEmitter.Reset()
		sinkManagerMetrics = metrics.NewSinkManagerMetrics()
	})

	AfterEach(func() {
		sinkManagerMetrics.Stop()
	})

	It("emits metrics for dump sinks", func() {
		sinkManagerMetrics := metrics.NewSinkManagerMetrics()
		Expect(fakeEventEmitter.GetMessages()).To(BeEmpty())

		sink = &dump.DumpSink{}
		sinkManagerMetrics.Inc(sink)

		expected := fake.Message{
			Origin: "doppler",
			Event: &events.ValueMetric{
				Name:  proto.String("messageRouter.numberOfDumpSinks"),
				Value: proto.Float64(1),
				Unit:  proto.String("sinks"),
			},
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))

		fakeEventEmitter.Reset()

		sinkManagerMetrics.Dec(sink)

		expected.Event = &events.ValueMetric{
			Name:  proto.String("messageRouter.numberOfDumpSinks"),
			Value: proto.Float64(0),
			Unit:  proto.String("sinks"),
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))
	})

	It("emits metrics for syslog sinks", func() {
		Expect(fakeEventEmitter.GetMessages()).To(BeEmpty())

		sink = &syslog.SyslogSink{}
		sinkManagerMetrics.Inc(sink)

		expected := fake.Message{
			Origin: "doppler",
			Event: &events.ValueMetric{
				Name:  proto.String("messageRouter.numberOfSyslogSinks"),
				Value: proto.Float64(1),
				Unit:  proto.String("sinks"),
			},
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))

		fakeEventEmitter.Reset()

		sinkManagerMetrics.Dec(sink)

		expected.Event = &events.ValueMetric{
			Name:  proto.String("messageRouter.numberOfSyslogSinks"),
			Value: proto.Float64(0),
			Unit:  proto.String("sinks"),
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))
	})

	It("emits metrics for websocket sinks", func() {
		Expect(fakeEventEmitter.GetMessages()).To(BeEmpty())

		sink = &websocket.WebsocketSink{}
		sinkManagerMetrics.Inc(sink)

		expected := fake.Message{
			Origin: "doppler",
			Event: &events.ValueMetric{
				Name:  proto.String("messageRouter.numberOfWebsocketSinks"),
				Value: proto.Float64(1),
				Unit:  proto.String("sinks"),
			},
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))

		fakeEventEmitter.Reset()

		sinkManagerMetrics.Dec(sink)

		expected.Event = &events.ValueMetric{
			Name:  proto.String("messageRouter.numberOfWebsocketSinks"),
			Value: proto.Float64(0),
			Unit:  proto.String("sinks"),
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))
	})

	It("emits metrics for firehose sinks", func() {
		Expect(fakeEventEmitter.GetMessages()).To(BeEmpty())

		sinkManagerMetrics.IncFirehose()

		expected := fake.Message{
			Origin: "doppler",
			Event: &events.ValueMetric{
				Name:  proto.String("messageRouter.numberOfFirehoseSinks"),
				Value: proto.Float64(1),
				Unit:  proto.String("sinks"),
			},
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))

		fakeEventEmitter.Reset()

		sinkManagerMetrics.DecFirehose()

		expected.Event = &events.ValueMetric{
			Name:  proto.String("messageRouter.numberOfFirehoseSinks"),
			Value: proto.Float64(0),
			Unit:  proto.String("sinks"),
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))
	})

	It("emits metrics for container metric sinks", func() {
		Expect(fakeEventEmitter.GetMessages()).To(BeEmpty())

		sink = &containermetric.ContainerMetricSink{}
		sinkManagerMetrics.Inc(sink)

		expected := fake.Message{
			Origin: "doppler",
			Event: &events.ValueMetric{
				Name:  proto.String("messageRouter.numberOfContainerMetricSinks"),
				Value: proto.Float64(1),
				Unit:  proto.String("sinks"),
			},
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))

		fakeEventEmitter.Reset()

		sinkManagerMetrics.Dec(sink)

		expected.Event = &events.ValueMetric{
			Name:  proto.String("messageRouter.numberOfContainerMetricSinks"),
			Value: proto.Float64(0),
			Unit:  proto.String("sinks"),
		}
		Eventually(fakeEventEmitter.GetMessages, 2).Should(ContainElement(expected))
	})
})
