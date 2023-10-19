package alerter_test

import (
	"bytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"kartverket.no/smseagle-proxy/pkg/alerter"
	"kartverket.no/smseagle-proxy/pkg/config"
	"kartverket.no/smseagle-proxy/pkg/smseagle"
	"net/http"
	"os"
)

func (mock *mockNotifier) Notify(message *smseagle.SMSEagleMessage) error {
	mock.notifyCalled = true
	mock.message = *message
	return nil
}

type mockNotifier struct {
	message      smseagle.SMSEagleMessage
	notifyCalled bool
}

var _ = Describe("Grafana", func() {
	cfg := config.ProxyConfig{}
	var server *ghttp.Server
	var grafana *alerter.Grafana
	var rawWebhook []byte
	var err error
	var mock mockNotifier
	var req *http.Request
	var client *http.Client

	// we start the server and prepare the notifier
	BeforeEach(func() {
		server = ghttp.NewServer()
		mock = mockNotifier{}
		grafana = alerter.NewGrafana(&mock, &cfg)
		client = &http.Client{}
	})
	// close server, reset structs
	AfterEach(func() {
		server.Close()
		mock = mockNotifier{}
		grafana = &alerter.Grafana{}
	})
	It("loads file correctly", func() {
		rawWebhook, err = os.ReadFile("../../test_files/grafana_webhooks/oncall_webhook.json")
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("Bad request", func() {
		BeforeEach(func() {
			req, err = http.NewRequest(http.MethodPost, server.URL()+"/webhook/sms", bytes.NewReader(rawWebhook))
			Expect(err).ShouldNot(HaveOccurred())
			server.RouteToHandler(http.MethodPost, "/webhook/sms", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPost, "/webhook/sms"),
				grafana.HandleSMS,
			))
		})
		It("returns bad request when team header is missing", func() {
			res, err := client.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})

	Describe("SMS request", func() {
		BeforeEach(func() {
			req, err = http.NewRequest(http.MethodPost, server.URL()+"/webhook/sms", bytes.NewReader(rawWebhook))
			Expect(err).ShouldNot(HaveOccurred())
			server.RouteToHandler(http.MethodPost, "/webhook/sms", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPost, "/webhook/sms"),
				grafana.HandleSMS,
			))
		})
		Context("Request for infrastrukturdrift is successful", func() {
			BeforeEach(func() {
				req.Header.Set("team", "infrastrukturdrift")
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(true))
			})
			It("should go to infrastrukturdrift", func() {
				Expect(mock.message.Receiver).Should(Equal(smseagle.Infrastrukturdrift))
			})
			It("message should be correct", func() {
				Expect(mock.message.Message).Should(Equal("Ny alarm: http://grafana:3000/a/grafana-oncall-app/alert-groups/I57917WDFNGHY"))
			})
			It("should have sms contact type", func() {
				Expect(mock.message.ContactType).Should(Equal(smseagle.SMS))
			})
		})
		Context("Request for appdrift is successful", func() {
			BeforeEach(func() {
				req.Header.Set("team", "appdrift")
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(true))
			})
			It("should go to appdrift", func() {
				Expect(mock.message.Receiver).Should(Equal(smseagle.Appdrift))
			})
			It("message should be correct", func() {
				Expect(mock.message.Message).Should(Equal("Ny alarm: http://grafana:3000/a/grafana-oncall-app/alert-groups/I57917WDFNGHY"))
			})
			It("should have sms contact type", func() {
				Expect(mock.message.ContactType).Should(Equal(smseagle.SMS))
			})
		})
	})
	Describe("Call request", func() {
		BeforeEach(func() {
			req, err = http.NewRequest(http.MethodPost, server.URL()+"/webhook/call", bytes.NewReader(rawWebhook))
			Expect(err).ShouldNot(HaveOccurred())
			server.RouteToHandler(http.MethodPost, "/webhook/call", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPost, "/webhook/call"),
				grafana.HandleCall,
			))
		})
		Context("Request for infrastrukturdrift is successful", func() {
			BeforeEach(func() {
				req.Header.Set("team", "infrastrukturdrift")
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(true))
			})
			It("should go to infrastrukturdrift", func() {
				Expect(mock.message.Receiver).Should(Equal(smseagle.Infrastrukturdrift))
			})
			It("message should be correct", func() {
				Expect(mock.message.Message).Should(Equal("Ny alarm: http://grafana:3000/a/grafana-oncall-app/alert-groups/I57917WDFNGHY"))
			})
			It("should have call contact type", func() {
				Expect(mock.message.ContactType).Should(Equal(smseagle.Call))
			})
		})
		Context("Request for appdrift is successful", func() {
			BeforeEach(func() {
				req.Header.Set("team", "appdrift")
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(true))
			})
			It("should go to appdrift", func() {
				Expect(mock.message.Receiver).Should(Equal(smseagle.Appdrift))
			})
			It("message should be correct", func() {
				Expect(mock.message.Message).Should(Equal("Ny alarm: http://grafana:3000/a/grafana-oncall-app/alert-groups/I57917WDFNGHY"))
			})
			It("should have call contact type", func() {
				Expect(mock.message.ContactType).Should(Equal(smseagle.Call))
			})
		})
	})
})
