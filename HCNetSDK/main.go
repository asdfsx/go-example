package main

/*
#include "HCNetSDK.h"
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
void exceptionCallBackFunc(DWORD ,LONG ,LONG ,void*);
void realDataCallBackV30(LONG lPlayHandle, DWORD dwDataType, BYTE *pBuffer, DWORD dwBufSize, void* pUser);
void standardDataCallBackV30(LONG lPlayHandle, DWORD dwDataType, BYTE *pBuffer, DWORD dwBufSize, void* pUser);
*/
import "C"
import (
	"errors"
	"fmt"
	"log"
	"time"
	"unsafe"
)

// 是否有错误
func isErr(oper string) error {
	errno := int64(C.NET_DVR_GetLastError())
	if errno > 0 {
		reMsg := fmt.Sprintf("%s摄像头失败,失败代码号：%d", oper, errno)
		return errors.New(reMsg)
	}
	return nil
}

// 初始化海康摄像头
func Init() (err error) {
	success := C.NET_DVR_Init()
	if success < 0 {
		if err = isErr("Init"); err != nil {
			return
		}
	}
	// 设置连接时间
	success = C.NET_DVR_SetConnectTime(C.DWORD(2000), C.DWORD(1))
	if success < 0 {
		if err = isErr("SetConnectTime"); err != nil {
			return
		}
	}
	c_logfile := C.CString("/tmp/hcnetsdk.log")
	defer C.free(unsafe.Pointer(c_logfile))

	C.NET_DVR_SetLogToFile(C.DWORD(3), c_logfile, C.BOOL(1))

	cryptoPath := C.CString("/vagrant_data/gocode/src/HCNetSDK/build/Linux/libcrypto.so")
	defer C.free(unsafe.Pointer(cryptoPath))
	C.NET_DVR_SetSDKInitCfg(C.NET_SDK_INIT_CFG_LIBEAY_PATH, unsafe.Pointer(cryptoPath))

	sslPath := C.CString("/vagrant_data/gocode/src/HCNetSDK/build/Linux/libssl.so")
	defer C.free(unsafe.Pointer(sslPath))
	C.NET_DVR_SetSDKInitCfg(C.NET_SDK_INIT_CFG_SSLEAY_PATH, unsafe.Pointer(sslPath))

	sdkPath := C.CString("/vagrant_data/gocode/src/HCNetSDK/build/Linux")
	defer C.free(unsafe.Pointer(sdkPath))
	var struComPath C.NET_DVR_LOCAL_SDK_PATH
	C.memcpy(unsafe.Pointer(&struComPath.sPath[0]), unsafe.Pointer(sdkPath), C.NET_SDK_MAX_FILE_PATH)
	C.NET_DVR_SetSDKInitCfg(C.NET_SDK_INIT_CFG_SDK_PATH, unsafe.Pointer(&struComPath))
	return nil
}

// 登录摄像头
func Login(username, password, addr string, port int) (int64, C.LPNET_DVR_DEVICEINFO_V30, error) {
	var deviceinfoV30 C.NET_DVR_DEVICEINFO_V30
	c_ip := C.CString(addr)
	defer C.free(unsafe.Pointer(c_ip))

	c_login := C.CString(username)
	defer C.free(unsafe.Pointer(c_login))

	c_password := C.CString(password)
	defer C.free(unsafe.Pointer(c_password))

	msgId := C.NET_DVR_Login_V30(c_ip, C.WORD(port), c_login, c_password,
		(C.LPNET_DVR_DEVICEINFO_V30)(unsafe.Pointer(&deviceinfoV30)),
	)

	if int64(msgId) < 0 {
		if err := isErr("Login"); err != nil {
			return -1, nil, err
		}
		return -1, nil, errors.New("登录摄像头失败")
	}

	fmt.Printf("ChanNum:%v, StartChan:%v, IPChanNum:%v, StartDChan:%v, HighDChanNum:%v\n", deviceinfoV30.byChanNum, deviceinfoV30.byStartChan, deviceinfoV30.byIPChanNum, deviceinfoV30.byStartDChan, deviceinfoV30.byHighDChanNum)
	fmt.Printf("MainProto: %v, SubProto: %d\n", deviceinfoV30.byMainProto, deviceinfoV30.bySubProto)

	return int64(msgId), &deviceinfoV30, nil
}

// 注册异常回调函数
func RegistExceptionCallback() error {
	issuccess := C.NET_DVR_SetExceptionCallBack_V30(C.UINT(0), nil, C.ExceptionCallBack(C.exceptionCallBackFunc), nil)
	if issuccess == 0 {
		if err := isErr("RegistExceptionCallback"); err != nil {
			return err
		}
		return errors.New("注册异常回调函数失败")
	}
	return nil
}

func GetDVRConfig(uid int64, deviceinfoV30 C.LPNET_DVR_DEVICEINFO_V30) map[string]int64 {
	var ipparacfg C.NET_DVR_IPPARACFG
	var returndLen C.DWORD
	var result map[string]int64 = map[string]int64{}
	ret := C.NET_DVR_GetDVRConfig(C.LONG(uid), C.NET_DVR_GET_IPPARACFG, C.LONG(0), (C.LPVOID)(unsafe.Pointer(&ipparacfg)), C.sizeof_NET_DVR_IPPARACFG, (C.LPDWORD)(unsafe.Pointer(&returndLen)))
	if ret > 0 { // ip camera，自己本身是个NVR、DVR
		// nvr/dvr 上模拟通道的设备
		// nvr/dvr 上模拟通道的设备
		for i := uint8(0); i < C.MAX_IP_DEVICE; i++ {
			ip := C.GoString(&ipparacfg.struIPDevInfo[i].struIP.sIpV4[0])
			fmt.Printf("device ipv4: %s\n", ip)
		}
		for i := uint8(0); i < C.MAX_IP_DEVICE; i++ {
			iDevID := int64(uint8(ipparacfg.struIPChanInfo[i].byIPIDHigh))*256 + int64(uint8(ipparacfg.struIPChanInfo[i].byIPID))
        	fmt.Printf("devid: %d, byIPIDHigh;%d, byIPID:%d,  channelid:%d\n", iDevID, ipparacfg.struIPChanInfo[i].byIPIDHigh, ipparacfg.struIPChanInfo[i].byIPID, ipparacfg.struIPChanInfo[i].byChannel)
        }
		for iChannum := uint8(0); iChannum < uint8(deviceinfoV30.byChanNum); iChannum++ {
			if ipparacfg.byAnalogChanEnable[iChannum] == 1 {
				key := fmt.Sprintf("Camara%04d", iChannum)
				result[key] = int64(iChannum + uint8(deviceinfoV30.byStartChan))
			}
		}
		// nvr/dvr 上数字通道的设备
		for iChannum := uint8(0); iChannum < uint8(deviceinfoV30.byIPChanNum); iChannum++ {
			if ipparacfg.struIPChanInfo[iChannum].byEnable == 1 {
				key := fmt.Sprintf("Camara%04d", iChannum)
				var deviceID int64 = int64(iChannum) + int64(uint8(deviceinfoV30.byStartDChan)) + int64(uint8(deviceinfoV30.byHighDChanNum))*256
				result[key] = deviceID
				fmt.Println(int64(iChannum), int64(uint8(deviceinfoV30.byStartDChan)), int64(uint8(deviceinfoV30.byHighDChanNum)))
			}
		}
	} else { // device not support,has no ip camera, 自己本身是个IPC设备
		fmt.Printf("ChanNum:%v, StartChan:%v, IPChanNum:%v\n", deviceinfoV30.byChanNum, deviceinfoV30.byStartChan, deviceinfoV30.byIPChanNum)
		for iChannum := uint8(0); iChannum < uint8(deviceinfoV30.byChanNum); iChannum++ {
			key := fmt.Sprintf("Camara%04d", iChannum)
			result[key] = int64(iChannum + uint8(deviceinfoV30.byStartChan))
		}
	}
	return result
}

func GetDVRConfig2(uid int64) error {
    var usercfg C.NET_DVR_USER
	var returndLen C.DWORD
	ret := C.NET_DVR_GetDVRConfig(C.LONG(uid), C.NET_DVR_GET_USERCFG, C.LONG(0), (C.LPVOID)(unsafe.Pointer(&usercfg)), C.sizeof_NET_DVR_USER, (C.LPDWORD)(unsafe.Pointer(&returndLen)))
	if ret > 0 {
		// dwsize := uint32(C.uint(usercfg.dwSize))
		// num := dwsize / C.sizeof_NET_DVR_USER_INFO
		// fmt.Println("------------------", num)
		for i := uint32(0);i< C.MAX_USERNUM; i++{
			fmt.Printf("%s, password:%s\n", usercfg.struUser[i].sUserName, usercfg.struUser[i].sPassword)
		}
	} else {
		if err := isErr("NET_DVR_GET_USERCFG"); err != nil {
			return err
		}
		return errors.New("查询用户失败")
	}


	return nil
}

// 登录摄像头
func Login_V40() (int64, error) {
	var deviceinfoV40 C.NET_DVR_DEVICEINFO_V40
	c_ip := C.CString("192.168.3.17")
	defer C.free(unsafe.Pointer(c_ip))

	c_login := C.CString("admin")
	defer C.free(unsafe.Pointer(c_login))

	c_password := C.CString("Asdf21852188")
	defer C.free(unsafe.Pointer(c_password))

	var c_logininfo C.NET_DVR_USER_LOGIN_INFO
	c_logininfo.wPort = C.WORD(8000)
	C.memcpy(unsafe.Pointer(&(c_logininfo.sDeviceAddress[0])), unsafe.Pointer(c_ip), C.NET_DVR_DEV_ADDRESS_MAX_LEN)
	C.memcpy(unsafe.Pointer(&(c_logininfo.sUserName[0])), unsafe.Pointer(c_login), C.NAME_LEN)
	C.memcpy(unsafe.Pointer(&(c_logininfo.sPassword[0])), unsafe.Pointer(c_password), C.NAME_LEN)

	msgId := C.NET_DVR_Login_V40((*C.NET_DVR_USER_LOGIN_INFO)(unsafe.Pointer(&c_logininfo)),
		(C.LPNET_DVR_DEVICEINFO_V40)(unsafe.Pointer(&deviceinfoV40)),
	)

	if int64(msgId) < 0 {
		if err := isErr("Login"); err != nil {
			return -1, err
		}
		return -1, errors.New("登录摄像头失败")
	}
	return int64(msgId), nil
}

// 退出摄像头登录
// uid:摄像头登录成功的id
func Logout(uid int64) error {
	C.NET_DVR_Logout_V30(C.LONG(uid))
	if err := isErr("Logout"); err != nil {
		return err
	}
	return nil
}

// 播放视频
// uid:摄像头登录成功的id
// 返回播放视频标识 pid
func Play(uid int64, channelID int64) (int64, error) {
	var pDetectInfo C.NET_DVR_CLIENTINFO
	pDetectInfo.lChannel = C.LONG(channelID)
	pid := C.NET_DVR_RealPlay_V30(C.LONG(uid), (C.LPNET_DVR_CLIENTINFO)(unsafe.Pointer(&pDetectInfo)), nil, nil, C.BOOL(1))
	if int64(pid) < 0 {
		if err := isErr("Play"); err != nil {
			return -1, err
		}
		return -1, errors.New("播放失败")
	}

	return int64(pid), nil
}

// 抓拍图片
func CaptureJpeg(uid int64, channelID int64) (string, error) {
	picPath := "/tmp/" + time.Now().Format("20060102150405") + ".jpeg"

	var jpegpara C.NET_DVR_JPEGPARA
	c_path := C.CString(picPath)
	defer C.free(unsafe.Pointer(c_path))
	msgId := C.NET_DVR_CaptureJPEGPicture(C.LONG(uid), C.LONG(channelID),
		(C.LPNET_DVR_JPEGPARA)(unsafe.Pointer(&jpegpara)),
		c_path,
	)

	if int64(msgId) < 0 {
		if err := isErr("Capture"); err != nil {
			return "", err
		}
		return "", errors.New("抓拍失败")
	}
	return picPath, nil
}

// 抓拍视频
func CaptureVideo(uid int64, channelID int64) (string, error) {
	now := time.Now().Format("20060102150405")
	videoPath := "/tmp/real_" + now + ".mp4"
	c_videopath := C.CString(videoPath)

	var pDetectInfo C.NET_DVR_CLIENTINFO
	pDetectInfo.lChannel = C.LONG(channelID)
	pid := C.NET_DVR_RealPlay_V30(C.LONG(uid), (C.LPNET_DVR_CLIENTINFO)(unsafe.Pointer(&pDetectInfo)), nil, nil, C.BOOL(1))
	if int64(pid) < 0 {
		if err := isErr("Play"); err != nil {
			return "", err
		}
		return "", errors.New("播放失败")
	}
	defer C.NET_DVR_StopRealPlay(pid)

	isuccess := C.NET_DVR_SaveRealData(pid, c_videopath)
	if isuccess == 0 {
		if err := isErr("SaveRealData"); err != nil {
			return "", err
		}
		return "", errors.New("预览中抓图失败")
	}
	defer C.NET_DVR_StopSaveRealData(pid)

	time.Sleep(10 * time.Second)

	return videoPath, nil
}

// 抓取数据流
func CaptureStream(uid, channelID int64) error {
	var struPlayInfo C.NET_DVR_PREVIEWINFO
	struPlayInfo.hPlayWnd     = C.uint(0)                        //需要SDK解码时句柄设为有效值，仅取流不解码时可设为空
    struPlayInfo.lChannel     = C.LONG(channelID)                //预览通道号
    struPlayInfo.dwStreamType = C.DWORD(0)                       //0-主码流，1-子码流，2-码流3，3-码流4，以此类推
    struPlayInfo.dwLinkMode   = C.DWORD(0)                       //0- TCP方式，1- UDP方式，2- 多播方式，3- RTP方式，4-RTP/RTSP，5-RSTP/HTTP
	struPlayInfo.bBlocked     = C.DWORD(0);                      //0- 非阻塞取流，1- 阻塞取流
	
	pid := C.NET_DVR_RealPlay_V40(C.LONG(uid), 
		(C.LPNET_DVR_PREVIEWINFO)(unsafe.Pointer(&struPlayInfo)), C.REALDATACALLBACK(C.realDataCallBackV30), nil)
	if int64(pid) < 0 {
		if err := isErr("CaptureStream"); err != nil {
			return err
		}
		return errors.New("抓取视频流失败")
	}
	defer C.NET_DVR_StopRealPlay(pid)
	time.Sleep(10 * time.Second)
	return nil
}

// 抓取标准数据流
func CaptureStandardStream(uid, channelID int64) error {
	// var struPlayInfo C.NET_DVR_PREVIEWINFO
	// struPlayInfo.hPlayWnd     = C.uint(0)                        //需要SDK解码时句柄设为有效值，仅取流不解码时可设为空
    // struPlayInfo.lChannel     = C.LONG(channelID)                //预览通道号
    // struPlayInfo.dwStreamType = C.DWORD(0)                       //0-主码流，1-子码流，2-码流3，3-码流4，以此类推
    // struPlayInfo.dwLinkMode   = C.DWORD(0)                       //0- TCP方式，1- UDP方式，2- 多播方式，3- RTP方式，4-RTP/RTSP，5-RSTP/HTTP
	// struPlayInfo.bBlocked     = C.DWORD(1);                      //0- 非阻塞取流，1- 阻塞取流
	
	// pid := C.NET_DVR_RealPlay_V40(C.LONG(uid), 
	// 	(C.LPNET_DVR_PREVIEWINFO)(unsafe.Pointer(&struPlayInfo)), nil, nil)
	// if int64(pid) < 0 {
	// 	if err := isErr("CaptureStream"); err != nil {
	// 		return err
	// 	}
	// 	return errors.New("抓取标准数据流失败")
	// }
	// defer C.NET_DVR_StopRealPlay(pid)

	// // issuccess := C.NET_DVR_SetRealDataCallBack(pid, C.STDDATACALLBACK(C.standardDataCallBackV30), C.DWORD(0))
	// issuccess := C.NET_DVR_SetStandardDataCallBack(pid, C.STDDATACALLBACK(C.standardDataCallBackV30), C.DWORD(0))
	// if issuccess == 0 {
	// 	if err := isErr("SetStandardDataCallBack"); err != nil {
	// 		return err
	// 	}
	// 	return errors.New("注册抓取标准数据流回调函数失败")
	// }

	// time.Sleep(10 * time.Second)
	// return nil

	var struClientInfo C.NET_DVR_CLIENTINFO 
	struClientInfo.hPlayWnd     = C.HWND(0);         //需要SDK解码时句柄设为有效值，仅取流不解码时可设为空
	struClientInfo.lChannel     = C.LONG(channelID);       //预览通道号
	struClientInfo.lLinkMode    = C.LONG(0);       //最高位(31)为0表示主码流，为1表示子码流0～30位表示连接方式：0－TCP方式；1－UDP方式；2－多播方式；3－RTP方式;
	struClientInfo.sMultiCastIP = nil;   //多播地址，需要多播预览时配置  
	
	pid := C.NET_DVR_RealPlay_V30(C.LONG(uid), 
		C.LPNET_DVR_CLIENTINFO(unsafe.Pointer(&struClientInfo)), nil, nil, C.BOOL(1))
	if int64(pid) < 0 {
		if err := isErr("CaptureStandardStream"); err != nil {
			return err
		}
		return errors.New("抓取标准数据流失败")
	}
	defer C.NET_DVR_StopRealPlay(pid)
	issuccess := C.NET_DVR_SetStandardDataCallBack(pid, C.STDDATACALLBACK(C.standardDataCallBackV30), C.DWORD(0))
	if issuccess == 0 {
		if err := isErr("SetStandardDataCallBack"); err != nil {
			return err
		}
		return errors.New("注册抓取标准数据流回调函数失败")
	}

	time.Sleep(10 * time.Second)
	return nil
}

// 操作云台
func PtzMove(uid int64, channelID int64) error {
	var pDetectInfo C.NET_DVR_CLIENTINFO
	pDetectInfo.lChannel = C.LONG(channelID)
	pid := C.NET_DVR_RealPlay_V30(C.LONG(uid), (C.LPNET_DVR_CLIENTINFO)(unsafe.Pointer(&pDetectInfo)), nil, nil, C.BOOL(1))
	if int64(pid) < 0 {
		if err := isErr("Play"); err != nil {
			return err
		}
		return errors.New("播放失败")
	}
	defer C.NET_DVR_StopRealPlay(pid)

	// 左
	isuccess := C.NET_DVR_PTZControl(pid, C.PAN_LEFT, 0)
	if isuccess == 0 {
		if err := isErr("PAN_LEFT"); err != nil {
			return err
		}
		return errors.New("视角左转失败")
	}
	defer C.NET_DVR_PTZControl(pid, C.PAN_LEFT, 1)

	time.Sleep(10 * time.Second)

	return nil
}

// 操作云台2
func PtzMove2(pid int64) error {
	// 右
	isuccess := C.NET_DVR_PTZControl(C.LONG(pid), C.PAN_RIGHT, 0)
	if isuccess == 0 {
		if err := isErr("PAN_RIGHT"); err != nil {
			return err
		}
		return errors.New("视角左转失败")
	}
	defer C.NET_DVR_PTZControl(C.LONG(pid), C.PAN_RIGHT, 1)

	time.Sleep(10 * time.Second)

	return nil
}

// 操作云台Other
func PtzMoveOther(uid int64, channelID int64) error {
	// 左
	isuccess := C.NET_DVR_PTZControl_Other(C.LONG(uid), C.LONG(channelID), C.PAN_LEFT, 0)
	if isuccess == 0 {
		if err := isErr("PAN_LEFT_OTHER"); err != nil {
			return err
		}
		return errors.New("视角左转失败")
	}
	time.Sleep(5 * time.Second)

	// 右
	isuccess = C.NET_DVR_PTZControl_Other(C.LONG(uid), C.LONG(channelID), C.PAN_RIGHT, 0)
	if isuccess == 0 {
		if err := isErr("PAN_RIGHT_OTHER"); err != nil {
			return err
		}
		return errors.New("视角右转失败")
	}
	time.Sleep(5 * time.Second)

	// 上
	isuccess = C.NET_DVR_PTZControl_Other(C.LONG(uid), C.LONG(channelID), C.TILT_UP, 0)
	if isuccess == 0 {
		if err := isErr("TILT_UP_OTHER"); err != nil {
			return err
		}
		return errors.New("视角上仰失败")
	}
	time.Sleep(5 * time.Second)

	isuccess = C.NET_DVR_PTZControl_Other(C.LONG(uid), C.LONG(channelID), C.TILT_DOWN, 0)
	if isuccess == 0 {
		if err := isErr("TILT_DOWN_OTHER"); err != nil {
			return err
		}
		return errors.New("视角下俯失败")
	}
	time.Sleep(5 * time.Second)
	defer C.NET_DVR_PTZControl_Other(C.LONG(uid), C.LONG(channelID), C.TILT_DOWN, 1)
	return nil
}

// 停止相机
// pid 播放标识符
func PtzStop(pid int64) error {
	msgId := C.NET_DVR_StopRealPlay(C.LONG(pid))
	if int64(msgId) < 0 {
		if err := isErr("PtzStop"); err != nil {
			return err
		}
		return errors.New("停止相机失败")
	}
	return nil
}

// 释放SDK资源，在程序结束之前调用。
func Close() {
	C.NET_DVR_Cleanup()
}

func main() {
	var err error
	err = Init()
	defer Close()
	if err != nil {
		log.Fatal(err.Error())
	}

	var uid int64
	var deviceinfoV30 C.LPNET_DVR_DEVICEINFO_V30
	// if uid, deviceinfoV30, err = Login("admin", "Asdf21852188", "192.168.3.17", 8000); err != nil {
	// 	log.Fatal(err.Error())
	// }

	if uid, deviceinfoV30, err = Login("admin", "Xingyun6666", "172.16.0.168", 8000); err != nil {
		log.Fatal(err.Error())
	}

	result := GetDVRConfig(uid, deviceinfoV30)
	var channelID int64
	for key, value := range result {
		fmt.Printf("channel %s:%d\n", key, value)
		channelID = value
	}

	if err = GetDVRConfig2(uid); err != nil{
		log.Fatal(err.Error())
	}

	var picPath string
	if picPath, err = CaptureJpeg(uid, channelID); err != nil {
		log.Fatal(err.Error())
	}
	log.Println("图片路径:", picPath)

	var videoPath string
	if videoPath, err = CaptureVideo(uid, channelID); err != nil {
		log.Fatal(err.Error())
	}
	log.Println("视频路径:", videoPath)

	if err = PtzMoveOther(uid, channelID); err != nil {
		log.Fatal(err.Error())
	}

	if err = PtzMove(uid, channelID); err != nil {
		log.Fatal(err.Error())
	}

	if err = CreateRealDataFile("/tmp/hcnet.psdat"); err != nil{
		log.Fatal(err.Error())
	}
	defer CloseRealDataFile()
	if err = CaptureStream(uid, channelID); err != nil{
		log.Fatal(err.Error())
	}
	log.Println("实时流文件路径: /tmp/hcnet.psdat")

	if err = CreateStandardFile("/tmp/hcnet.stdat"); err != nil{
		log.Fatal(err.Error())
	}
	defer CloseStandardFile()
	if err = CaptureStandardStream(uid, channelID); err != nil{
		log.Fatal(err.Error())
	}
	log.Println("标准流文件路径: /tmp/hcnet.stdat")

	var pid int64
	if pid, err = Play(uid, channelID); err != nil {
		log.Fatal(err.Error())
	}

	if err = PtzMove2(pid); err != nil {
		log.Fatal(err.Error())
	}

	if err = PtzStop(pid); err != nil {
		log.Fatal(err.Error())
	}

	if err = Logout(uid); err != nil {
		log.Fatal(err.Error())
	}

}
