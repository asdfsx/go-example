package goav_test

import (
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avdevice"
	"github.com/giorgisio/goav/avfilter"
	"github.com/giorgisio/goav/avformat"
	"github.com/giorgisio/goav/avutil"
	"github.com/giorgisio/goav/swresample"
	"github.com/giorgisio/goav/swscale"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Goav", func() {
	Context("using command", func() {
		It("test lib", func() {
			avformat.AvRegisterAll()
			avcodec.AvcodecRegisterAll()

			log.Printf("AvFilter Version:\t%v", avfilter.AvfilterVersion())
			log.Printf("AvDevice Version:\t%v", avdevice.AvdeviceVersion())
			log.Printf("SWScale Version:\t%v", swscale.SwscaleVersion())
			log.Printf("AvUtil Version:\t%v", avutil.AvutilVersion())
			log.Printf("AvCodec Version:\t%v", avcodec.AvcodecVersion())
			log.Printf("Resample Version:\t%v", swresample.SwresampleLicense())
		})
		It("test mp4", func() {
			// Open video file
			pFormatContext := avformat.AvformatAllocContext()
			if avformat.AvformatOpenInput(&pFormatContext, videoAddr, nil, nil) != 0 {
				Fail(fmt.Sprintf("Unable to open file %s\n", videoAddr))
			}
			defer pFormatContext.AvformatCloseInput()

			// Retrieve stream information
			if pFormatContext.AvformatFindStreamInfo(nil) < 0 {
				Fail("Couldn't find stream information")
			}

			// Dump information about file onto standard error
			pFormatContext.AvDumpFormat(0, videoAddr, 0)
		})
		It("test rtsp pull and save frame", func() {
			pFormatContext := avformat.AvformatAllocContext()

			if errcode := avformat.AvformatOpenInput(&pFormatContext, rtspAddr, nil, nil); errcode < 0 {
				Fail(fmt.Sprintf("Unable to open rtsp: %s, cause: %s", rtspAddr, avutil.ErrorFromCode(errcode)))
			}

			if errcode := pFormatContext.AvformatFindStreamInfo(nil); errcode < 0 {
				pFormatContext.AvformatCloseInput()
				Fail(fmt.Sprintf("Couldn't find stream information, cause: %s", avutil.ErrorFromCode(errcode)))
			}

			for i := 0; i < int(pFormatContext.NbStreams()); i++ {
				switch pFormatContext.Streams()[i].CodecParameters().AvCodecGetType() {
				case avformat.AVMEDIA_TYPE_VIDEO:
					// Get a pointer to the codec context for the video stream
					pCodecCtxOrig := pFormatContext.Streams()[i].Codec()
					// Find the decoder for the video stream
					pCodec := avcodec.AvcodecFindDecoder(avcodec.CodecId(pCodecCtxOrig.GetCodecId()))
					if pCodec == nil {
						Fail("Unsupported codec!")
					}
					// Copy context
					pCodecCtx := pCodec.AvcodecAllocContext3()
					if errcode := pCodecCtx.AvcodecCopyContext((*avcodec.Context)(unsafe.Pointer(pCodecCtxOrig))); errcode < 0 {
						Fail(fmt.Sprintf("Couldn't copy codec context, cause: %s", avutil.ErrorFromCode(errcode)))
					}

					// Open codec
					if errcode := pCodecCtx.AvcodecOpen2(pCodec, nil); errcode < 0 {
						Fail(fmt.Sprintf("Could not open codec, cause: %s", avutil.ErrorFromCode(errcode)))
					}

					// Allocate video frame
					pFrame := avutil.AvFrameAlloc()

					// Allocate an AVFrame structure
					pFrameRGB := avutil.AvFrameAlloc()
					if pFrameRGB == nil {
						Fail("Unable to allocate RGB Frame")
					}

					// Determine required buffer size and allocate buffer
					numBytes := uintptr(avcodec.AvpictureGetSize(avcodec.AV_PIX_FMT_RGB24, pCodecCtx.Width(),
						pCodecCtx.Height()))
					buffer := avutil.AvMalloc(numBytes)

					// Assign appropriate parts of buffer to image planes in pFrameRGB
					// Note that pFrameRGB is an AVFrame, but AVFrame is a superset
					// of AVPicture
					avp := (*avcodec.Picture)(unsafe.Pointer(pFrameRGB))
					avp.AvpictureFill((*uint8)(buffer), avcodec.AV_PIX_FMT_RGB24, pCodecCtx.Width(), pCodecCtx.Height())

					// initialize SWS context for software scaling
					swsCtx := swscale.SwsGetcontext(
						pCodecCtx.Width(),
						pCodecCtx.Height(),
						(swscale.PixelFormat)(pCodecCtx.PixFmt()),
						pCodecCtx.Width(),
						pCodecCtx.Height(),
						avcodec.AV_PIX_FMT_RGB24,
						avcodec.SWS_BILINEAR,
						nil,
						nil,
						nil,
					)

					// Read frames and save first five frames to disk
					frameNumber := 1
					packet := avcodec.AvPacketAlloc()
					for pFormatContext.AvReadFrame(packet) >= 0 {
						// Is this a packet from the video stream?
						if packet.StreamIndex() == i {
							// Decode video frame
							response := pCodecCtx.AvcodecSendPacket(packet)
							if response < 0 {
								fmt.Printf("Error while sending a packet to the decoder: %s\n", avutil.ErrorFromCode(response))
							}
							for response >= 0 {
								response = pCodecCtx.AvcodecReceiveFrame((*avcodec.Frame)(unsafe.Pointer(pFrame)))
								if response == avutil.AvErrorEAGAIN || response == avutil.AvErrorEOF {
									break
								} else if response < 0 {
									fmt.Printf("Error while receiving a frame from the decoder: %s\n", avutil.ErrorFromCode(response))
									return
								}

								if frameNumber <= 5 {
									// Convert the image from its native format to RGB
									swscale.SwsScale2(swsCtx, avutil.Data(pFrame),
										avutil.Linesize(pFrame), 0, pCodecCtx.Height(),
										avutil.Data(pFrameRGB), avutil.Linesize(pFrameRGB))

									// Save the frame to disk
									fmt.Printf("Writing frame %d\n", frameNumber)
									_ = SaveFrame(pFrameRGB, pCodecCtx.Width(), pCodecCtx.Height(), frameNumber)
								} else {
									return
								}
								frameNumber++
							}
						}

						// Free the packet that was allocated by av_read_frame
						packet.AvFreePacket()
					}

					// Free the RGB image
					avutil.AvFree(buffer)
					avutil.AvFrameFree(pFrameRGB)

					// Free the YUV frame
					avutil.AvFrameFree(pFrame)

					// Close the codecs
					pCodecCtx.AvcodecClose()
					(*avcodec.Context)(unsafe.Pointer(pCodecCtxOrig)).AvcodecClose()

					// Close the video file
					pFormatContext.AvformatCloseInput()

					// Stop after saving frames of first video straem
					break

				default:
					Fail("Didn't find a video stream")
				}
			}
		})
		It("test rtps pull and only save jpeg", func() {
			pFormatContext := avformat.AvformatAllocContext()

			if errcode := avformat.AvformatOpenInput(&pFormatContext, rtspAddr, nil, nil); errcode < 0 {
				Fail(fmt.Sprintf("Unable to open rtsp: %s, cause: %s", rtspAddr, avutil.ErrorFromCode(errcode)))
			}

			if errcode := pFormatContext.AvformatFindStreamInfo(nil); errcode < 0 {
				pFormatContext.AvformatCloseInput()
				Fail(fmt.Sprintf("Couldn't find stream information, cause: %s", avutil.ErrorFromCode(errcode)))
			}

			for i := 0; i < int(pFormatContext.NbStreams()); i++ {
				switch pFormatContext.Streams()[i].CodecParameters().AvCodecGetType() {
				case avformat.AVMEDIA_TYPE_VIDEO:
					// Get a pointer to the codec context for the video stream
					pCodecCtxOrig := pFormatContext.Streams()[i].Codec()
					// Find the decoder for the video stream
					pCodec := avcodec.AvcodecFindDecoder(avcodec.CodecId(pCodecCtxOrig.GetCodecId()))
					if pCodec == nil {
						Fail("Unsupported codec!")
					}
					// Copy context
					pCodecCtx := pCodec.AvcodecAllocContext3()
					if errcode := pCodecCtx.AvcodecCopyContext((*avcodec.Context)(unsafe.Pointer(pCodecCtxOrig))); errcode < 0 {
						Fail(fmt.Sprintf("Couldn't copy codec context, cause: %s", avutil.ErrorFromCode(errcode)))
					}

					// Open codec
					if errcode := pCodecCtx.AvcodecOpen2(pCodec, nil); errcode < 0 {
						Fail(fmt.Sprintf("Could not open codec, cause: %s", avutil.ErrorFromCode(errcode)))
					}

					// Allocate video frame
					pFrame := avutil.AvFrameAlloc()

					frameNumber := 1
					packet := avcodec.AvPacketAlloc()
					for pFormatContext.AvReadFrame(packet) >= 0 {
						response := pCodecCtx.AvcodecSendPacket(packet)
						if response < 0 {
							fmt.Printf("Error while sending a packet to the decoder: %s\n", avutil.ErrorFromCode(response))
						}
						for response >= 0 {
							response = pCodecCtx.AvcodecReceiveFrame((*avcodec.Frame)(unsafe.Pointer(pFrame)))

							if response == avutil.AvErrorEAGAIN || response == avutil.AvErrorEOF {
								break
							} else if response < 0 {
								fmt.Printf("Error while receiving a frame from the decoder: %s\n", avutil.ErrorFromCode(response))
								return
							}
							if frameNumber <= 5 {
								fmt.Printf("Writing frame %d\n", frameNumber)
								err := SaveFrameToJpeg(pFrame, frameNumber)
								Expect(err).NotTo(HaveOccurred())
							} else {
								return
							}
							frameNumber++
						}

						// Free the packet that was allocated by av_read_frame
						packet.AvFreePacket()
					}

					// Free the YUV frame
					avutil.AvFrameFree(pFrame)

					// Close the codecs
					pCodecCtx.AvcodecClose()
					(*avcodec.Context)(unsafe.Pointer(pCodecCtxOrig)).AvcodecClose()

					// Close the video file
					pFormatContext.AvformatCloseInput()

					// Stop after saving frames of first video straem
					break
				default:
					Fail("Didn't find a video stream")
				}
			}
		})
		It("test rtsp pull and only save video", func() {
			var ifcx *avformat.Context
			var iccx *avcodec.Context
			var ist *avformat.Stream
			var i_index int
			var timenow, timestart time.Time

			var ofcx *avformat.Context
			var ofmt *avformat.OutputFormat
			var occx *avcodec.Context
			var ost *avformat.Stream

			var pkt *avcodec.Packet

			var ix int64

			//
			// Input
			//

			// open rtsp
			if errcode := avformat.AvformatOpenInput(&ifcx, rtspAddr, nil, nil); errcode < 0 {
				Fail(fmt.Sprintf("Unable to open rtsp: %s, cause: %s", rtspAddr, avutil.ErrorFromCode(errcode)))
			}

			if errcode := ifcx.AvformatFindStreamInfo(nil); errcode < 0 {
				ifcx.AvformatCloseInput()
				Fail(fmt.Sprintf("ERROR: Cannot find stream info: %s", avutil.ErrorFromCode(errcode)))
			}

			// search video stream
			i_index = -1
			for ix := 0; ix < int(ifcx.NbStreams()); ix++ {
				if ifcx.Streams()[ix].CodecParameters().AvCodecGetType() == avformat.AVMEDIA_TYPE_VIDEO {
					iccx = (*avcodec.Context)(unsafe.Pointer(ifcx.Streams()[ix].Codec()))
					ist = ifcx.Streams()[ix]
					i_index = ix
					break
				}
			}
			if i_index < 0 {
				ifcx.AvformatCloseInput()
				Fail("ERROR: Cannot find input video stream")
			}

			//
			// Output
			//

			// open output file
			ofmt = avformat.AvGuessFormat("", videoOutput, "")
			if errcode := avformat.AvformatAllocOutputContext2(&ofcx, ofmt, "", videoOutput); errcode < 0 {
				Fail(fmt.Sprintf("Couldn't allocate output, cause: %s", avutil.ErrorFromCode(errcode)))
			}
			pb, err := avformat.AvIOOpen(videoOutput, avformat.AVIO_FLAG_WRITE)
			Expect(err).NotTo(HaveOccurred())
			ofcx.SetPb(pb)

			// Create output stream
			// ost = avformat_new_stream( ofcx, (AVCodec *) iccx->codec );
			ost = ofcx.AvformatNewStream(nil)
			occx = (*avcodec.Context)(unsafe.Pointer(ost.Codec()))
			if errcode := occx.AvcodecCopyContext((*avcodec.Context)(unsafe.Pointer(iccx))); errcode < 0 {
				Fail(fmt.Sprintf("Couldn't copy codec context, cause: %s", avutil.ErrorFromCode(errcode)))
			}

			ost.AvStreamSetRFrameRate(ist.AvStreamGetRFrameRate())

			fmt.Printf("----------input codec timebase: %v----------", ist.Codec().GetTimeBase())
			fmt.Printf("----------output stream timebase: %v--------", ost.Codec().GetTimeBase())

			ost.AvStreamGetRFrameRate()

			if errcode := ofcx.AvformatWriteHeader(nil); errcode < 0 {
				Fail(fmt.Sprintf("Failed to write header, cause: %s", avutil.ErrorFromCode(errcode)))
			}
			ifcx.AvDumpFormat(0, ifcx.Filename(), 0)
			ofcx.AvDumpFormat(0, ofcx.Filename(), 1)

			timenow = time.Now()
			timestart = timenow

			ix = 0
			// av_read_play(context);//play RTSP (Shouldn't need this since it defaults to playing on connect)
			pkt = avcodec.AvPacketAlloc()

			for {
				if errcode := ifcx.AvReadFrame(pkt); errcode < 0 {
					if errcode < 0 {
						fmt.Println("error while reading frames: ", avutil.ErrorFromCode(errcode))
						break
					}
				}

				if pkt.StreamIndex() != i_index { // packet is not video
					continue
				}

				// Make sure we start on a key frame
				if timestart == timenow && !(pkt.Flags() == avcodec.AV_PKT_FLAG_KEY) {
					timenow = time.Now()
					timestart = timenow
					continue
				}

				logPacket(ost, pkt)

				pkt.SetStreamIndex(ost.Id())

				pkt.SetPts(ix)
				pkt.SetDts(ix)
				ix++

				ofcx.AvInterleavedWriteFrame(pkt)

				pkt.AvFreePacket()
				pkt.AvInitPacket()

				timenow = time.Now()
			}

			iccx.AvcodecClose()

			ifcx.AvReadPause()
			ifcx.AvformatFreeContext()

			ofcx.AvWriteTrailer()
			ofcx.Pb().Close()
			ofcx.AvformatFreeContext()
		})
		It("test rtsp pull and save video & ppm", func() {
			var ifcx *avformat.Context
			var ifccx *avformat.CodecContext
			var iccx *avcodec.Context
			var ist *avformat.Stream
			var ic *avcodec.Codec

			var iccxForDecode *avcodec.Context

			var i_index int
			var timenow, timestart time.Time

			var ofcx *avformat.Context
			var ofmt *avformat.OutputFormat
			var occx *avcodec.Context
			var ost *avformat.Stream

			var pkt *avcodec.Packet

			// Allocate video frame
			var frame *avutil.Frame
			var frameRGB *avutil.Frame

			var ix int64

			//
			// Input
			//

			// open rtsp
			if errcode := avformat.AvformatOpenInput(&ifcx, rtspAddr, nil, nil); errcode < 0 {
				Fail(fmt.Sprintf("Unable to open rtsp: %s, cause: %s", rtspAddr, avutil.ErrorFromCode(errcode)))
			}

			if errcode := ifcx.AvformatFindStreamInfo(nil); errcode < 0 {
				ifcx.AvformatCloseInput()
				Fail(fmt.Sprintf("ERROR: Cannot find stream info: %s", avutil.ErrorFromCode(errcode)))
			}

			// search video stream
			i_index = -1
			for ix := 0; ix < int(ifcx.NbStreams()); ix++ {
				if ifcx.Streams()[ix].CodecParameters().AvCodecGetType() == avformat.AVMEDIA_TYPE_VIDEO {
					ifccx = ifcx.Streams()[ix].Codec()
					iccx = (*avcodec.Context)(unsafe.Pointer(ifccx))
					ist = ifcx.Streams()[ix]
					i_index = ix
					break
				}
			}
			if i_index < 0 {
				ifcx.AvformatCloseInput()
				Fail("ERROR: Cannot find input video stream")
			}

			// Find the decoder for the video stream
			ic = avcodec.AvcodecFindDecoder(avcodec.CodecId(ifccx.GetCodecId()))
			if ic == nil {
				Fail("Unsupported codec!")
			}

			// Copy context
			iccxForDecode = ic.AvcodecAllocContext3()
			if errcode := iccxForDecode.AvcodecCopyContext(iccx); errcode < 0 {
				Fail(fmt.Sprintf("Couldn't copy codec context, cause: %s", avutil.ErrorFromCode(errcode)))
			}

			// Open codec
			if errcode := iccxForDecode.AvcodecOpen2(ic, nil); errcode < 0 {
				Fail(fmt.Sprintf("Could not open codec, cause: %s", avutil.ErrorFromCode(errcode)))
			}

			//
			// Output
			//

			// open output file
			ofmt = avformat.AvGuessFormat("", videoOutput, "")
			if errcode := avformat.AvformatAllocOutputContext2(&ofcx, ofmt, "", videoOutput); errcode < 0 {
				Fail(fmt.Sprintf("Couldn't allocate output, cause: %s", avutil.ErrorFromCode(errcode)))
			}
			pb, err := avformat.AvIOOpen(videoOutput, avformat.AVIO_FLAG_WRITE)
			Expect(err).NotTo(HaveOccurred())
			ofcx.SetPb(pb)

			// Create output stream
			// ost = avformat_new_stream( ofcx, (AVCodec *) iccx->codec );
			ost = ofcx.AvformatNewStream(nil)
			occx = (*avcodec.Context)(unsafe.Pointer(ost.Codec()))
			if errcode := occx.AvcodecCopyContext((*avcodec.Context)(unsafe.Pointer(iccx))); errcode < 0 {
				Fail(fmt.Sprintf("Couldn't copy codec context, cause: %s", avutil.ErrorFromCode(errcode)))
			}

			ost.AvStreamSetRFrameRate(ist.AvStreamGetRFrameRate())

			fmt.Printf("----------input codec: %v----------\n", avcodec.AvcodecGetName(avcodec.CodecId(ist.Codec().GetCodecId())))
			fmt.Printf("----------output codec: %v----------\n", avcodec.AvcodecGetName(avcodec.CodecId(ost.Codec().GetCodecId())))
			fmt.Printf("----------input codec timebase: %v---------\n", ist.Codec().GetTimeBase())
			fmt.Printf("----------output codec timebase: %v--------\n", ost.Codec().GetTimeBase())

			ost.AvStreamGetRFrameRate()

			if errcode := ofcx.AvformatWriteHeader(nil); errcode < 0 {
				Fail(fmt.Sprintf("Failed to write header, cause: %s", avutil.ErrorFromCode(errcode)))
			}
			ifcx.AvDumpFormat(0, ifcx.Filename(), 0)
			ofcx.AvDumpFormat(0, ofcx.Filename(), 1)

			timenow = time.Now()
			timestart = timenow

			ix = 0
			// av_read_play(context);//play RTSP (Shouldn't need this since it defaults to playing on connect)
			pkt = avcodec.AvPacketAlloc()
			frame = avutil.AvFrameAlloc()

			// Allocate an AVFrame structure
			frameRGB = avutil.AvFrameAlloc()
			if frameRGB == nil {
				Fail("Unable to allocate RGB Frame")
			}

			// Determine required buffer size and allocate buffer
			numBytes := uintptr(avcodec.AvpictureGetSize(avcodec.AV_PIX_FMT_RGB24, iccx.Width(),
				iccx.Height()))
			buffer := avutil.AvMalloc(numBytes)

			// Assign appropriate parts of buffer to image planes in pFrameRGB
			// Note that pFrameRGB is an AVFrame, but AVFrame is a superset
			// of AVPicture
			avp := (*avcodec.Picture)(unsafe.Pointer(frameRGB))
			avp.AvpictureFill((*uint8)(buffer), avcodec.AV_PIX_FMT_RGB24, iccx.Width(), iccx.Height())

			// initialize SWS context for software scaling
			swsCtx := swscale.SwsGetcontext(
				iccx.Width(),
				iccx.Height(),
				(swscale.PixelFormat)(iccx.PixFmt()),
				iccx.Width(),
				iccx.Height(),
				avcodec.AV_PIX_FMT_RGB24,
				avcodec.SWS_BILINEAR,
				nil,
				nil,
				nil,
			)

			for {
				if errcode := ifcx.AvReadFrame(pkt); errcode < 0 {
					if errcode < 0 {
						fmt.Println("error while reading frames: ", avutil.ErrorFromCode(errcode))
						break
					}
				}

				if pkt.StreamIndex() != i_index { // packet is not video
					continue
				}

				// Make sure we start on a key frame
				if timestart == timenow && !(pkt.Flags() == avcodec.AV_PKT_FLAG_KEY) {
					timenow = time.Now()
					timestart = timenow
					continue
				}

				logPacket(ost, pkt)

				response := iccxForDecode.AvcodecSendPacket(pkt)
				if response < 0 {
					fmt.Printf("Error while sending a packet to the decoder: %s\n", avutil.ErrorFromCode(response))
				} else {
					response = iccxForDecode.AvcodecReceiveFrame((*avcodec.Frame)(unsafe.Pointer(frame)))
					if response == avutil.AvErrorEAGAIN || response == avutil.AvErrorEOF {
						fmt.Printf("Error while receiving a frame from the decoder: %s\n", avutil.ErrorFromCode(response))
					} else if response < 0 {
						fmt.Printf("Error while receiving a frame from the decoder: %s\n", avutil.ErrorFromCode(response))
					} else {
						if ix%10 == 0 {
							// Convert the image from its native format to RGB
							errcode := swscale.SwsScale2(swsCtx, avutil.Data(frame),
								avutil.Linesize(frame), 0, iccx.Height(),
								avutil.Data(frameRGB), avutil.Linesize(frameRGB))
							fmt.Printf("Error while convert a frame to RGB: %s\n", avutil.ErrorFromCode(errcode))

							err := SaveFrame(frameRGB, iccx.Width(), iccx.Height(), int(ix))
							Expect(err).NotTo(HaveOccurred())
						}
					}
				}

				pkt.SetStreamIndex(ost.Id())

				pkt.SetPts(ix)
				pkt.SetDts(ix)
				ix++

				ofcx.AvInterleavedWriteFrame(pkt)

				pkt.AvFreePacket()
				pkt.AvInitPacket()

				timenow = time.Now()
			}

			avutil.AvFrameFree(frame)
			avutil.AvFrameFree(frameRGB)

			iccx.AvcodecClose()
			iccxForDecode.AvcodecClose()

			ifcx.AvReadPause()
			ifcx.AvformatFreeContext()

			ofcx.AvWriteTrailer()
			ofcx.Pb().Close()
			ofcx.AvformatFreeContext()
		})
	})
})

// SaveFrame writes a single frame to disk as a PPM file
func SaveFrame(frame *avutil.Frame, width, height, frameNumber int) error {
	// Open file
	fileName := fmt.Sprintf("frame%d.ppm", frameNumber)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	header := fmt.Sprintf("P6\n%d %d\n255\n", width, height)
	_, _ = file.Write([]byte(header))

	// Write pixel data
	for y := 0; y < height; y++ {
		data0 := avutil.Data(frame)[0]
		buf := make([]byte, width*3)
		startPos := uintptr(unsafe.Pointer(data0)) + uintptr(y)*uintptr(avutil.Linesize(frame)[0])
		for i := 0; i < width*3; i++ {
			element := *(*uint8)(unsafe.Pointer(startPos + uintptr(i)))
			buf[i] = element
		}
		_, _ = file.Write(buf)
	}
	return nil
}

// SaveFrame writes a single frame to disk as a PPM file
func SaveFrameToJpeg(frame *avutil.Frame, frameNumber int) error {
	// Open file
	fileName := fmt.Sprintf("frame%d.jpg", frameNumber)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := avutil.GetPicture(frame)
	if err != nil {
		return err
	}

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return err
	}
	return nil
}

func logPacket(streamCtx *avformat.Stream, pkt *avcodec.Packet) {
	timeBase := streamCtx.TimeBase()

	fmt.Printf("pts:%v pts_time:%v dts:%v dts_time:%v duration:%v duration_time:%v stream_index:%v\n",
		pkt.Pts(), pkt.Pts()*int64(timeBase.Num())%int64(timeBase.Den()),
		pkt.Dts(), pkt.Dts()*int64(timeBase.Num())%int64(timeBase.Den()),
		pkt.Duration(), pkt.Duration()*timeBase.Num()%timeBase.Den(),
		pkt.StreamIndex())
}
