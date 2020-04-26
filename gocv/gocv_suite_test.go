package gocv_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGocv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gocv Suite")
}
