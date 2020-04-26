package goav_test

import (
	"testing"

	"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avdevice"
	"github.com/giorgisio/goav/avfilter"
	"github.com/giorgisio/goav/avformat"
	"github.com/giorgisio/goav/avutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoav(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Goav Suite")
}

var (
	videoAddr   string = "/Users/sunxia/gocode/src/git.in-clustar.ai/steel/detect-center/example/goav/test.mp4"
	rtspAddr    string = "rtsp://localhost/test"
	videoOutput string = "/Users/sunxia/gocode/src/git.in-clustar.ai/steel/detect-center/example/goav/test2.mp4"
)

var _ = BeforeSuite(func() {
	// ffmpeg := "/usr/local/bin/ffmpeg"
	// params := []string{"-re","-i", videoAddr, "-rtsp_transport", "tcp", "-vcodec", "h264", "-f", "rtsp", rtspAddr}
	// cmd := exec.Command(ffmpeg, params...)
	// go func(){
	// 	err := cmd.Start()
	// 	fmt.Println(err)
	// }()
	avdevice.AvdeviceRegisterAll()
	avfilter.AvfilterRegisterAll()
	avformat.AvRegisterAll()
	avcodec.AvcodecRegisterAll()
	avformat.AvformatNetworkInit()
	avutil.AvLogSetLevel(avutil.AV_LOG_TRACE)
})

var _ = AfterSuite(func() {
	avformat.AvformatNetworkDeinit()
})
