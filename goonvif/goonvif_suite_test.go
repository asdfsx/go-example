package goonvif_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rikugun/goonvif"
)

func TestGoonvif(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Goonvif Suite")
}

var (
	camaraURL = "192.168.3.17"
	camaraPort = 80
	camaraUser = "operator"
	camaraPass = "Tekken31234"

	dev *goonvif.OnvifDevice
	err error
)

var _ = BeforeSuite(func(){
	dev, err = goonvif.NewDevice(fmt.Sprintf("%s:%d", camaraURL, camaraPort))
	Expect(err).NotTo(HaveOccurred())
	dev.Authenticate(camaraUser, camaraPass)
})

var _ = AfterSuite(func(){
})