package shiny_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFoo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shiny Suite")
}
