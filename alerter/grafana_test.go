package alerter_test

import (
	"bytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"kartverket.no/smseagle-proxy/alerter"
	"kartverket.no/smseagle-proxy/notifier"
	"net/http"
	"os"
)

func (mock *mockNotifier) Notify(message *notifier.SMSEagleMessage) error {
	mock.message = *message
	return nil
}

type mockNotifier struct {
	message notifier.SMSEagleMessage
}

var _ = Describe("Grafana", func() {

	var server *ghttp.Server
	var grafana *alerter.Grafana
	var rawWebhook []byte
	var err error
	var mock mockNotifier

	// we start the server and prepare the notifier
	BeforeEach(func() {
		server = ghttp.NewServer()
		mock = mockNotifier{}
		grafana = alerter.NewGrafana(&mock)
		server.AppendHandlers(grafana.HandleWebhook)
	})
	// close server, reset structs
	AfterEach(func() {
		server.Close()
		mock = mockNotifier{}
		grafana = &alerter.Grafana{}
	})

	Describe("Non-critical multi alert", func() {
		It("loads file correctly", func() {
			rawWebhook, err = os.ReadFile("../test_files/grafana_webhooks/non_critical_multi_alert.json")
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("Request is successful", func() {
			BeforeEach(func() {
				res, err := http.Post(server.URL()+"/webhook", "application/json", bytes.NewReader(rawWebhook))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})

			It("should call", func() {
				Expect(mock.message.Call).Should(Equal(false))
			})
			It("should go to appdrift", func() {
				Expect(mock.message.Receiver).Should(Equal(notifier.Appdrift))
			})
			It("message should be correct", func() {
				Expect(mock.message.Message).Should(Equal("[FIRING:6] skyline Test (istiod http-monitoring pilot true)"))
			})
		})
	})
})
