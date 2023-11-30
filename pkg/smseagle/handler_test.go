package smseagle_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"kartverket.no/smseagle-proxy/pkg/config"
	. "kartverket.no/smseagle-proxy/pkg/smseagle"
	"net/http"
)

var _ = Describe("Handler", func() {
	var cfg config.ProxyConfig
	var smseagle *SMSEagle
	var server *ghttp.Server

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfg = config.ProxyConfig{
			Call: config.SMSEagleConfig{
				Url:         server.URL(),
				AccessToken: "calltoken",
			},
			SMS: config.SMSEagleConfig{
				Url:         server.URL(),
				AccessToken: "smstoken",
			},
		}
		smseagle = NewSMSEagle(&cfg)
	})
	AfterEach(func() {
		smseagle = &SMSEagle{}
		cfg = config.ProxyConfig{}
		server.Close()
	})

	Context("appdrift alert", func() {
		It("should make get requests with correct queries to sms and call", func() {
			msg := SMSEagleMessage{Message: "hei pa deg, urltest: https://kartverket.atlassian.net/wiki/spaces/SKIP/pages/713359536/Playbook+for+SKIP-alarmer#HostOutOfInodes", PhoneNumber: "123"}
			exptectedSMSMsg := "hei+pa+deg,+urltest:+https://kartverket.atlassian.net/wiki/spaces/SKIP/pages/713359536/Playbook%2Bfor%2BSKIP-alarmer%23HostOutOfInodes"
			expectedSMSQuery := fmt.Sprintf("access_token=%s&to=%s&message=%s&unicode=1", cfg.SMS.AccessToken, "123", exptectedSMSMsg)
			expectedCallQuery := fmt.Sprintf("access_token=%s&to=%s", cfg.Call.AccessToken, "123")
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/http_api/send_sms", expectedSMSQuery),
					ghttp.RespondWith(http.StatusOK, "OK"),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/http_api/call_with_termination", expectedCallQuery),
					ghttp.RespondWith(http.StatusOK, "OK"),
				),
			)
			err := smseagle.Notify(&msg)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

})
