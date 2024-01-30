package safe

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMorbydSafe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "morbyd/safe package")
}
