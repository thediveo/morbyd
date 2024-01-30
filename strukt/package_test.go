package strukt

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMorbydStrukt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "morbyd/strukt package")
}
