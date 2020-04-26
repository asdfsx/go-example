package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
	"fmt"
	"time"
	"os"
	"log"
)

//export exceptionCallBackFunc
func exceptionCallBackFunc(dwType C.uint, lUserID C.int, lHandle C.int, pUser unsafe.Pointer) {
    switch dwType {
    	case 0x8005:    //预览时重连
    		fmt.Printf("----------reconnect--------%v\n", time.Now())
    		break;
		default:
    		break;
    }
}

var pFile *os.File
var standardFile *os.File
var err error

func CreateRealDataFile(filename string) error {
	pFile, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		return nil
}

func CloseRealDataFile() error {
	return pFile.Close()
}

func CreateStandardFile(filename string) error {
	standardFile, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	return nil
}

func CloseStandardFile() error {
	return standardFile.Close()
}

//export realDataCallBackV30
func realDataCallBackV30(lRealHandle C.int, dwDataType C.uint, pBuffer *C.uchar, dwBufSize C.uint, dwUser unsafe.Pointer){
	if dwDataType == 1 {//系统头
		mySlice := C.GoBytes(unsafe.Pointer(pBuffer), C.int(dwBufSize))
		_, err = pFile.Write(mySlice)
		if err != nil{
			log.Println(err.Error())
		}
		// fmt.Printf("write head len=%d\n", dwBufSize);
	} else { //视频数据流
		mySlice := C.GoBytes(unsafe.Pointer(pBuffer), C.int(dwBufSize))
		_, err = pFile.Write(mySlice)
		if err != nil{
			log.Println(err.Error())
		}
		// fmt.Printf("write body len=%d\n", dwBufSize);
	}
}

//export standardDataCallBackV30
func standardDataCallBackV30(lRealHandle C.int, dwDataType C.uint, pBuffer *C.uchar, dwBufSize C.uint, dwUser unsafe.Pointer){
	if dwDataType == 1 {//系统头
		mySlice := C.GoBytes(unsafe.Pointer(pBuffer), C.int(dwBufSize))
		_, err = standardFile.Write(mySlice)
		if err != nil{
			log.Println(err.Error())
		}
	} else if dwDataType == 4 { //标准视频数据流
		mySlice := C.GoBytes(unsafe.Pointer(pBuffer), C.int(dwBufSize))
		_, err = standardFile.Write(mySlice)
		if err != nil{
			log.Println(err.Error())
		}
	}
}