package goonvif_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/beevik/etree"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/yakovlevdmv/goonvif/Device"
	"github.com/yakovlevdmv/goonvif/Media"
	"github.com/yakovlevdmv/goonvif/PTZ"
	"github.com/yakovlevdmv/goonvif/xsd"
	"github.com/yakovlevdmv/goonvif/xsd/onvif"
)

var _ = Describe("Goonvif", func() {
	Context("goonvif test", func(){
		It("call GetSystemDateAndTime", func(){
			m_tm_req := Device.GetSystemDateAndTime{}
			m_tm_resp, err := dev.CallMethod(m_tm_req)
			Expect(err).NotTo(HaveOccurred())
			tmp := readResponse(m_tm_resp)
			element := tmp.FindElement("./Envelope/Body/GetSystemDateAndTimeResponse/SystemDateAndTime/DateTimeType")
			fmt.Println(element.Text())
		})
		It("call GetProfiles", func(){
			m_media_profiles_req := Media.GetProfiles{}
			m_media_profiles_resp, err := dev.CallMethod(m_media_profiles_req)
			Expect(err).NotTo(HaveOccurred())
			tmp := readResponse(m_media_profiles_resp)
			elements := tmp.FindElements("./Envelope/Body/GetProfilesResponse/Profiles")
			for _, e := range elements {
				for _, a := range e.Attr{
					fmt.Println(a.Key, a.Value)
				}
			}
		})
		It("call GetStreamUri", func(){
			m_media_profiles_req := Media.GetStreamUri{
				ProfileToken:"Profile_1",
			}
			m_media_profiles_resp, err := dev.CallMethod(m_media_profiles_req)
			Expect(err).NotTo(HaveOccurred())
			tmp := readResponse(m_media_profiles_resp)
			e := tmp.FindElement("./Envelope/Body/GetStreamUriResponse/MediaUri")
			for _, e := range e.ChildElements() {
				fmt.Printf("%s : %s\n", e.Tag, e.Text())
			}
		})
		It("call PTZ.GetStatus", func(){
			m_ptz_stat_req := PTZ.GetStatus{
				ProfileToken:"Profile_1",
			}
			m_ptz_stat_resp, err := dev.CallMethod(m_ptz_stat_req)
			Expect(err).NotTo(HaveOccurred())
			tmp := readResponse(m_ptz_stat_resp)
			e := tmp.FindElement("./Envelope/Body/GetStatusResponse/PTZStatus")
			for _, e := range e.ChildElements() {
				if e.Tag == "Position" {
					fmt.Println("Position:")
					for _, e1 := range e.ChildElements(){
						switch e1.Tag {
						case "PanTilt":
							for _, a := range e1.Attr {
								fmt.Printf("\tPanTilt: %s: %s\n", a.Key, a.Value)
							}
						case "Zoom":
							for _, a := range e1.Attr {
								fmt.Printf("\tZoom: %s: %s\n", a.Key, a.Value)
							}
						}
					}
				}
				fmt.Printf("%s : %s\n", e.Tag, e.Text())
			}
		})
		It("call PTZ.RelativeMove", func(){
			m_ptz_rel_req := PTZ.RelativeMove{
				ProfileToken: "Profile_1",
				Translation: onvif.PTZVector{
					PanTilt: onvif.Vector2D{
						X:     -1.0,
						Y:     -1.0,
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace"),
					},
					Zoom: onvif.Vector1D{
						X:     -1.0,  // -1.0 -> 1.0
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/TranslationGenericSpace"),
					},
				},
				Speed: onvif.PTZSpeed{
					PanTilt: onvif.Vector2D{
						X:     0.0,
						Y:     0.0,
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/GenericSpeedSpace"),
					},
					Zoom: onvif.Vector1D{
						X:     0.0,
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/ZoomGenericSpeedSpace"),
					},
				},
			}
			m_ptz_rel_resp, err := dev.CallMethod(m_ptz_rel_req)
			if err != nil {
				fmt.Println("resp error!", err)
				return
			}
			tmp := readResponseRaw(m_ptz_rel_resp)
			fmt.Println(tmp)
		})
		It("call PTZ.GetStatus", func(){
			time.Sleep(5 * time.Second)
			m_ptz_stat_req := PTZ.GetStatus{
				ProfileToken:"Profile_1",
			}
			m_ptz_stat_resp, err := dev.CallMethod(m_ptz_stat_req)
			Expect(err).NotTo(HaveOccurred())
			tmp := readResponse(m_ptz_stat_resp)
			e := tmp.FindElement("./Envelope/Body/GetStatusResponse/PTZStatus")
			for _, e := range e.ChildElements() {
				if e.Tag == "Position" {
					fmt.Println("Position:")
					for _, e1 := range e.ChildElements(){
						switch e1.Tag {
						case "PanTilt":
							for _, a := range e1.Attr {
								fmt.Printf("\tPanTilt: %s: %s\n", a.Key, a.Value)
							}
						case "Zoom":
							for _, a := range e1.Attr {
								fmt.Printf("\tZoom: %s: %s\n", a.Key, a.Value)
							}
						}
					}
				}
				fmt.Printf("%s : %s\n", e.Tag, e.Text())
			}
		})
		It("call PTZ.AbsoluteMove", func(){
			time.Sleep(1 * time.Second)
				m_ptz_abs_req := PTZ.AbsoluteMove{
				ProfileToken: "Profile_1",
				Position: onvif.PTZVector{
					PanTilt: onvif.Vector2D{
						X:     -0.5,
						Y:     -0.5,
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace"),
					},
					Zoom: onvif.Vector1D{
						X:     0.0,  // 0.0 -> 1.0
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace"),
					},
				},
				Speed: onvif.PTZSpeed{
					PanTilt: onvif.Vector2D{
						X:     0.0,
						Y:     0.0,
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/GenericSpeedSpace"),
					},
					Zoom: onvif.Vector1D{
						X:     0.0,
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/ZoomGenericSpeedSpace"),
					},
				},
			}
			m_ptz_abs_resp, err := dev.CallMethod(m_ptz_abs_req)
			if err != nil {
				fmt.Println("resp error!", err)
				return
			}
			tmp := readResponseRaw(m_ptz_abs_resp)
			fmt.Println(tmp)
		})
		It("call PTZ.GetStatus", func(){
			time.Sleep(5 * time.Second)
			m_ptz_stat_req := PTZ.GetStatus{
				ProfileToken:"Profile_1",
			}
			m_ptz_stat_resp, err := dev.CallMethod(m_ptz_stat_req)
			Expect(err).NotTo(HaveOccurred())
			tmp := readResponse(m_ptz_stat_resp)
			e := tmp.FindElement("./Envelope/Body/GetStatusResponse/PTZStatus")
			for _, e := range e.ChildElements() {
				if e.Tag == "Position" {
					fmt.Println("Position:")
					for _, e1 := range e.ChildElements(){
						switch e1.Tag {
						case "PanTilt":
							for _, a := range e1.Attr {
								fmt.Printf("\tPanTilt: %s: %s\n", a.Key, a.Value)
							}
						case "Zoom":
							for _, a := range e1.Attr {
								fmt.Printf("\tZoom: %s: %s\n", a.Key, a.Value)
							}
						}
					}
				}
				fmt.Printf("%s : %s\n", e.Tag, e.Text())
			}
		})
		It("call PTZ.GotoPreset", func(){
			m_ptz_preset_req := PTZ.GotoPreset{
				ProfileToken: "Profile_1",
				PresetToken:  "2",
				Speed: onvif.PTZSpeed{
					PanTilt: onvif.Vector2D{
						X:     0.0,
						Y:     0.0,
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/GenericSpeedSpace"),
					},
					Zoom: onvif.Vector1D{
						X:     0.0,
						Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/ZoomGenericSpeedSpace"),
					},
				},
			}
			m_ptz_preset_resp, err := dev.CallMethod(m_ptz_preset_req)
			Expect(err).NotTo(HaveOccurred())
			tmp := readResponseRaw(m_ptz_preset_resp)
			fmt.Println(tmp)
		})
	})
})

func readResponse(resp *http.Response) *etree.Document {
	doc := etree.NewDocument()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Fail(err.Error())
	}

	if err := doc.ReadFromBytes(b); err != nil {
		Fail(err.Error())
	}
	return doc
}

func readResponseRaw(resp *http.Response) string {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Fail(err.Error())
	}
	return string(b)
}