package gocv_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	"gocv.io/x/gocv"
)

var (
	rtspAddr    string = "rtsp://localhost/test"
	videoOutput string = "/Users/sunxia/gocode/src/git.in-clustar.ai/steel/detect-center/example/gocv/test2.mp4"
	jpegOutput  string = "/Users/sunxia/gocode/src/git.in-clustar.ai/steel/detect-center/example/gocv/test2.jpeg"
)

var _ = Describe("Gocv", func() {
	Context("test rtsp", func() {
		It("pull rtsp", func() {
			rtsp, err := gocv.VideoCaptureFile(rtspAddr)
			if err != nil {
				fmt.Printf("Error opening video capture device: %v\n", rtspAddr)
				return
			}
			defer rtsp.Close()

			img := gocv.NewMat()
			defer img.Close()

			if ok := rtsp.Read(&img); !ok {
				fmt.Printf("cannot read rtsp %v\n", rtspAddr)
				return
			}

			writer, err := gocv.VideoWriterFile(videoOutput, "h264", 25, img.Cols(), img.Rows(), true)
			if err != nil {
				fmt.Printf("error opening video writer device: %v\n", videoOutput)
				return
			}
			defer writer.Close()

			for i := 0; i < 10000; i++ {
				if ok := rtsp.Read(&img); !ok {
					fmt.Printf("rtsp closed: %v\n", rtspAddr)
					return
				}
				if img.Empty() {
					continue
				}

				_ = writer.Write(img)
			}
		})
		It("pull image", func() {
			rtsp, err := gocv.VideoCaptureFile(rtspAddr)
			if err != nil {
				fmt.Printf("Error opening video capture device: %v\n", rtspAddr)
				return
			}
			defer rtsp.Close()

			img := gocv.NewMat()
			defer img.Close()

			if ok := rtsp.Read(&img); !ok {
				fmt.Printf("cannot read rtsp %v\n", rtspAddr)
				return
			}

			if img.Empty() {
				fmt.Printf("no image on rtsp %v\n", rtspAddr)
				return
			}

			gocv.IMWrite(jpegOutput, img)
		})
	})
})
