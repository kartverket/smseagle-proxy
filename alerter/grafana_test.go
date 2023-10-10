package alerter_test

import (
	"bytes"
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"kartverket.no/smseagle-proxy/alerter"
	"net/http"
	"os"
)

var _ = Describe("Grafana", func() {
	Describe("Non-critical multi alert - Should not call", func() {

		rawWebhook, err := os.ReadFile("../test_files/grafana_webhooks/non_critical_multi_alert.json")
		Expect(err).NotTo(HaveOccurred())
		webhook := alerter.GrafanaWebhook{}
		err = json.Unmarshal(rawWebhook, &webhook)
		Expect(err).NotTo(HaveOccurred())

		Context("When request is sent", func() {
			server := ghttp.NewServer()
			server.AppendHandlers(alerter.HandleWebhook)
			It("returns correct result", func() {
				res, err := http.Post(server.URL()+"/webhook", "application/json", bytes.NewReader(rawWebhook))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).Should(Equal(http.StatusOK))
			})
		})

	})
})
