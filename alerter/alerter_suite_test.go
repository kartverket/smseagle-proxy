package alerter_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAlerter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Alerter Suite")
}
