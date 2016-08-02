package logcounter_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"tools/logcounterapp/config"
	"tools/logcounterapp/logcounter"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logCounter", func() {
	Describe("Start", func() {
		It("returns 200 status for dumpReport", func() {
			port := testPort()

			cc := newMockCC()
			uaa := newMockUAA()
			cfg := &config.Config{
				ApiURL:         "api.test.com",
				DopplerURL:     "doppler.test.com",
				UaaURL:         "uaa.test.com",
				ClientID:       "testID",
				ClientSecret:   "clientSecret",
				Username:       "testUserName",
				Password:       "testPassword",
				MessagePrefix:  "testPrefix",
				SubscriptionID: "testSubID",
				Port:           port,
			}

			lc := logcounter.New(uaa, cc, cfg)
			go func() {
				err := lc.Start()
				Expect(err).ToNot(HaveOccurred())
			}()
			defer lc.Stop()

			Eventually(func() bool { return checkEndpoint(port, "?report", http.StatusOK) }).Should(BeTrue())
		})

		It("returns 404 status for unknown routes", func() {
			port := testPort()

			cc := newMockCC()
			uaa := newMockUAA()
			cfg := &config.Config{
				ApiURL:         "api.test.com",
				DopplerURL:     "doppler.test.com",
				UaaURL:         "uaa.test.com",
				ClientID:       "testID",
				ClientSecret:   "clientSecret",
				Username:       "testUserName",
				Password:       "testPassword",
				MessagePrefix:  "testPrefix",
				SubscriptionID: "testSubID",
				Port:           port,
			}

			lc := logcounter.New(uaa, cc, cfg)
			go func() {
				err := lc.Start()
				Expect(err).ToNot(HaveOccurred())
			}()
			defer lc.Stop()

			Eventually(func() bool { return checkEndpoint(port, "doesntExist", http.StatusNotFound) }).Should(BeTrue())

		})
	})

	Describe("Stop", func() {
		It("stops accepting requests", func() {
			port := testPort()

			cc := newMockCC()
			uaa := newMockUAA()
			cfg := &config.Config{
				ApiURL:         "api.test.com",
				DopplerURL:     "doppler.test.com",
				UaaURL:         "uaa.test.com",
				ClientID:       "testID",
				ClientSecret:   "clientSecret",
				Username:       "testUserName",
				Password:       "testPassword",
				MessagePrefix:  "testPrefix",
				SubscriptionID: "testSubID",
				Port:           port,
			}

			lc := logcounter.New(uaa, cc, cfg)
			go func() {
				err := lc.Start()
				Expect(err).ToNot(HaveOccurred())
			}()

			Eventually(func() bool { return checkEndpoint(port, "?report", http.StatusOK) }).Should(BeTrue())

			lc.Stop()
			Eventually(func() bool { return checkEndpoint(port, "?report", http.StatusOK) }).Should(BeFalse())
		})
	})

	Describe("HandleMessages", func() {
		It("results in a correct report when envelopes are passed to it", func() {
			port := testPort()

			cc := newMockCC()
			uaa := newMockUAA()
			cfg := &config.Config{
				ApiURL:         "api.test.com",
				DopplerURL:     "doppler.test.com",
				UaaURL:         "uaa.test.com",
				ClientID:       "testID",
				ClientSecret:   "clientSecret",
				Username:       "testUserName",
				Password:       "testPassword",
				MessagePrefix:  "testPrefix",
				SubscriptionID: "testSubID",
				Port:           port,
			}

			lc := logcounter.New(uaa, cc, cfg)
			go func() {
				err := lc.Start()
				Expect(err).ToNot(HaveOccurred())
			}()

			Eventually(func() bool { return checkEndpoint(port, "?report", http.StatusOK) }).Should(BeTrue())

			msgs := make(chan *events.Envelope, 10)
			for i := 0; i < 10; i++ {
				envelope := &events.Envelope{
					Origin:    proto.String("testOrigin"),
					EventType: events.Envelope_LogMessage.Enum(),
					LogMessage: &events.LogMessage{
						Message:     []byte(fmt.Sprintf("testPrefix guid: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx msg: %d", i)),
						MessageType: events.LogMessage_OUT.Enum(),
						Timestamp:   proto.Int64(time.Now().UnixNano()),
					},
				}

				msgs <- envelope
			}

			go lc.HandleMessages(msgs)
			Eventually(func() bool { return checkMessageBody(port, "?report", "total: 10") }).Should(BeTrue())
			lc.Stop()
		})

		It("doesn't process events that aren't logMessages", func() {
			port := testPort()

			cc := newMockCC()
			uaa := newMockUAA()
			cfg := &config.Config{
				ApiURL:         "api.test.com",
				DopplerURL:     "doppler.test.com",
				UaaURL:         "uaa.test.com",
				ClientID:       "testID",
				ClientSecret:   "clientSecret",
				Username:       "testUserName",
				Password:       "testPassword",
				MessagePrefix:  "testPrefix",
				SubscriptionID: "testSubID",
				Port:           port,
			}

			lc := logcounter.New(uaa, cc, cfg)
			go func() {
				err := lc.Start()
				Expect(err).ToNot(HaveOccurred())
			}()

			Eventually(func() bool { return checkEndpoint(port, "?report", http.StatusOK) }).Should(BeTrue())

			msgs := make(chan *events.Envelope, 1)

			counterEnvelope := &events.Envelope{
				Origin:    proto.String("testOrigin"),
				EventType: events.Envelope_CounterEvent.Enum(),
				LogMessage: &events.LogMessage{
					Message:     []byte("testPrefix guid: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx msg: 0"),
					MessageType: events.LogMessage_OUT.Enum(),
					Timestamp:   proto.Int64(time.Now().UnixNano()),
				},
			}
			msgs <- counterEnvelope

			go lc.HandleMessages(msgs)
			Eventually(func() bool { return checkMessageBody(port, "?report", "No messages received") }).Should(BeTrue())
			lc.Stop()
		})

		It("correctly counts the number of 1008 and other errors", func() {
			port := testPort()

			cc := newMockCC()
			uaa := newMockUAA()
			cfg := &config.Config{
				ApiURL:         "api.test.com",
				DopplerURL:     "doppler.test.com",
				UaaURL:         "uaa.test.com",
				ClientID:       "testID",
				ClientSecret:   "clientSecret",
				Username:       "testUserName",
				Password:       "testPassword",
				MessagePrefix:  "testPrefix",
				SubscriptionID: "testSubID",
				Port:           port,
			}

			lc := logcounter.New(uaa, cc, cfg)
			go func() {
				err := lc.Start()
				Expect(err).ToNot(HaveOccurred())
			}()

			Eventually(func() bool { return checkEndpoint(port, "?report", http.StatusOK) }).Should(BeTrue())

			terminate := make(chan os.Signal, 1)
			errs := make(chan error, 11)
			testErr := errors.New("Some Error")
			testWebsocketErr := errors.New("websocket: close 1008 Client did not respond to ping before keep-alive timeout expired.")

			closer := newMockCloser()
			close(closer.CloseOutput.Ret0)

			for i := 0; i < 2; i++ {
				errs <- testErr
				lc.HandleErrors(errs, terminate, closer)
			}

			for i := 0; i < 11; i++ {
				errs <- testWebsocketErr
				lc.HandleErrors(errs, terminate, closer)
			}

			Eventually(func() bool { return checkMessageBody(port, "?report", "1008 errors: 11") }).Should(BeTrue())
			Eventually(func() bool { return checkMessageBody(port, "?report", "Other errors: 2") }).Should(BeTrue())
			lc.Stop()
		})
	})
})

func testPort() string {
	add, _ := net.ResolveTCPAddr("tcp", ":0")
	l, _ := net.ListenTCP("tcp", add)
	defer l.Close()
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	return port
}

func checkEndpoint(port, endpoint string, status int) bool {
	resp, _ := http.Get("http://localhost:" + port + "/" + endpoint)
	if resp != nil {
		return resp.StatusCode == status
	}

	return false
}

func checkMessageBody(port, endpoint, expected string) bool {
	resp, _ := http.Get("http://localhost:" + port + "/" + endpoint)
	if resp != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		if strings.Contains(string(body), expected) {
			return true
		}
	}

	return false
}
