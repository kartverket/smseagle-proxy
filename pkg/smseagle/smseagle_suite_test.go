package smseagle_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSmseagle(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Smseagle Suite")
}
