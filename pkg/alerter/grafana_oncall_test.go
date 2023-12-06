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
	mock.timesCalled++
	return nil
}

type mockNotifier struct {
	message      smseagle.SMSEagleMessage
	notifyCalled bool
	timesCalled  int
}

var _ = Describe("GrafanaOncall", func() {
	cfg := config.ProxyConfig{}
	var server *ghttp.Server
	var grafana *alerter.GrafanaOncall
	var rawWebhook []byte
	var err error
	var mock mockNotifier
	var req *http.Request
	var client *http.Client

	// we start the server and prepare the notifier
	BeforeEach(func() {
		server = ghttp.NewServer()
		mock = mockNotifier{}
		grafana = alerter.NewGrafanaOncall(&mock, &cfg)
		client = &http.Client{}
	})
	// close server, reset structs
	AfterEach(func() {
		server.Close()
		mock = mockNotifier{}
		grafana = &alerter.GrafanaOncall{}
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
		It("returns bad request when phonenumber header is missing", func() {
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
		Context("Request for admin username is successful", func() {
			BeforeEach(func() {
				cfg.Users = map[string]string{"admin": "123"}
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(true))
				Expect(mock.timesCalled).Should(Equal(1))
			})
			It("should go to 123", func() {
				Expect(mock.message.PhoneNumber).Should(Equal("123"))
			})
			It("message should be correct", func() {
				Expect(mock.message.Message).Should(Equal("Ny Alarm \nId: I57917WDFNGHY \nOpprettet: 2023-10-12 12:17:12 \nTittel: [firing:3] InstanceDown  \nAntall: 1\nLenke: http://grafana:3000/a/grafana-oncall-app/alert-groups/I57917WDFNGHY\nPlaybook: https://kartverket.atlassian.net/wiki/spaces/SKIP/pages/713359536/Playbook+for+SKIP-alarmer#HostOutOfInodes"))
			})
			It("should have sms contact type", func() {
				Expect(mock.message.ContactType).Should(Equal(smseagle.SMS))
			})
		})
		Context("Request sms is correct when no playbook is provided", func() {
			BeforeEach(func() {
				rawWebhook, err = os.ReadFile("../../test_files/grafana_webhooks/oncall_webhook_no_playbook.json")
				Expect(err).ShouldNot(HaveOccurred())
				cfg.Users = map[string]string{"admin": "456"}
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(true))
			})
			It("should go to 456", func() {
				Expect(mock.message.PhoneNumber).Should(Equal("456"))
			})
			It("message should be correct", func() {
				Expect(mock.message.Message).Should(Equal("Ny Alarm \nId: I57917WDFNGHY \nOpprettet: 2023-10-12 12:17:12 \nTittel: [firing:3] InstanceDown  \nAntall: 1\nLenke: http://grafana:3000/a/grafana-oncall-app/alert-groups/I57917WDFNGHY"))
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
		Context("Request for 123 is successful", func() {
			BeforeEach(func() {
				cfg.Users = map[string]string{"admin": "123"}
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(true))
			})
			It("should go to 123", func() {
				Expect(mock.message.PhoneNumber).Should(Equal("123"))
			})
			It("should have call contact type", func() {
				Expect(mock.message.ContactType).Should(Equal(smseagle.Call))
			})
		})
	})
	Describe("Resolve SMS request", func() {
		BeforeEach(func() {
			rawWebhook, err = os.ReadFile("../../test_files/grafana_webhooks/oncall_resolved_webhook.json")
			Expect(err).ShouldNot(HaveOccurred())
			req, err = http.NewRequest(http.MethodPost, server.URL()+"/webhook/sms", bytes.NewReader(rawWebhook))
			Expect(err).ShouldNot(HaveOccurred())
			server.RouteToHandler(http.MethodPost, "/webhook/sms", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPost, "/webhook/sms"),
				grafana.HandleSMS,
			))
		})
		Context("Request for 123 is successful", func() {
			BeforeEach(func() {
				cfg.Users = map[string]string{"admin": "123"}
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(true))
			})
			It("should go to 123", func() {
				Expect(mock.message.PhoneNumber).Should(Equal("123"))
			})
			It("message should be correct", func() {
				Expect(mock.message.Message).Should(Equal("Alarm løst \nId: IAXB4WC5DVD9R \nLøst: 2023-10-24 11:9:11 \nTittel: [firing:3] InstanceDown  \nAntall: 1 \nLenke: http://grafana:3000/a/grafana-oncall-app/alert-groups/IAXB4WC5DVD9R"))
			})
			It("should have sms contact type", func() {
				Expect(mock.message.ContactType).Should(Equal(smseagle.SMS))
			})
		})
	})
	Describe("Escalation custom message request", func() {
		BeforeEach(func() {
			rawWebhook, err = os.ReadFile("../../test_files/grafana_webhooks/oncall_custom_message_webhook.json")
			Expect(err).ShouldNot(HaveOccurred())
			req, err = http.NewRequest(http.MethodPost, server.URL()+"/webhook/sms", bytes.NewReader(rawWebhook))
			Expect(err).ShouldNot(HaveOccurred())
			server.RouteToHandler(http.MethodPost, "/webhook/sms", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPost, "/webhook/sms"),
				grafana.HandleSMS,
			))
		})
		Context("Request for 123 is successful", func() {
			BeforeEach(func() {
				cfg.Users = map[string]string{"SomekartverkUser@kartverket.no": "299938423"}
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(true))
			})
			It("should go to 123", func() {
				Expect(mock.message.PhoneNumber).Should(Equal("299938423"))
			})
			It("message should be correct", func() {
				Expect(mock.message.Message).Should(Equal("%f0%9f%9a%a8New Alert!%f0%9f%9a%91\nTitle: [firing:3] InstanceDown\nID: I1TRVMRRPT31L\nCreated: 2023-12-04T12:54:36.741831Z\nPlaybook: https://kartverket.atlassian.net/wiki/spaces/SKIP/pages/713359536/Playbook+for+SKIP-alarmer#HostOutOfInodes"))
			})
			It("should have sms contact type", func() {
				Expect(mock.message.ContactType).Should(Equal(smseagle.SMS))
			})
		})
		Context("Handle no user to be notified correctly", func() {
			BeforeEach(func() {
				rawWebhook, err = os.ReadFile("../../test_files/grafana_webhooks/oncall_custom_message_webhook_no_user.json")
				Expect(err).ShouldNot(HaveOccurred())
				req, err = http.NewRequest(http.MethodPost, server.URL()+"/webhook/sms", bytes.NewReader(rawWebhook))
				Expect(err).ShouldNot(HaveOccurred())
				server.RouteToHandler(http.MethodPost, "/webhook/sms", ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/webhook/sms"),
					grafana.HandleSMS,
				))
				cfg.Users = map[string]string{"SomekartverkUser@kartverket.no": "299938423"}
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
			})
			It("should call notify", func() {
				Expect(mock.notifyCalled).Should(Equal(false))
			})
		})
		Context("Handle multiple users to be notified correctly", func() {
			BeforeEach(func() {
				rawWebhook, err = os.ReadFile("../../test_files/grafana_webhooks/oncall_custom_message_webhook_multiple_users.json")
				Expect(err).ShouldNot(HaveOccurred())
				req, err = http.NewRequest(http.MethodPost, server.URL()+"/webhook/sms", bytes.NewReader(rawWebhook))
				Expect(err).ShouldNot(HaveOccurred())
				server.RouteToHandler(http.MethodPost, "/webhook/sms", ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/webhook/sms"),
					grafana.HandleSMS,
				))
				cfg.Users = map[string]string{"somekartverkUser@kartverket.no": "299938423", "SomekartverkUser2@kartverket.no": "299938421"}
				res, err := client.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
			It("should call notify twice", func() {
				Expect(mock.timesCalled).Should(Equal(2))
			})
		})
	})
})
