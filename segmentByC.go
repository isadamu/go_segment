package main

/*

#cgo windows CFLAGS: -ID:/ffmpeg/include
#cgo windows LDFLAGS: -LD:/ffmpeg/lib -llibavformat  -llibavcodec -llibavutil -llibavdevice -llibavfilter -llibswresample -llibswscale

#cgo linux CFLAGS: -I/usr/local/include
#cgo linux LDFLAGS: -L/usr/local/lib -lavformat  -lavcodec -lavutil -lavdevice -lavfilter -lswresample -lswscale

#include <stdlib.h>
#include "segment.h"
*/
import "C"

import (
	"strings"
	"unsafe"
)

const (
	timeInterval = 20
	wrapLimit    = 3
)

type CSegment struct {
	taskId       string
	inputUrl     string
	outputFolder string
}

// 初始化
// void InitSnapshotStruct(void** ppSnapshot, char* taskId, char* inputUrl, char* outputFolder, int timeInterval, int wrapLimit);
func NewCSegment(taskId, inputUrl, outputFolder string) *CSegment {
	if strings.HasSuffix(outputFolder, "/") {
		outputFolder = outputFolder[:len(outputFolder)-1]
	}

	cs := &CSegment{
		taskId:       taskId,
		inputUrl:     inputUrl,
		outputFolder: outputFolder,
	}

	return cs
}

// 开始截图
func (cs *CSegment) Start() {
	taskIdC := C.CString(cs.taskId)
	inputUrlC := C.CString(cs.inputUrl)
	outputFolderC := C.CString(cs.outputFolder)

	defer C.free(unsafe.Pointer(taskIdC))
	defer C.free(unsafe.Pointer(inputUrlC))
	defer C.free(unsafe.Pointer(outputFolderC))

	timeInterval := C.int(timeInterval)
	wrapLimit := C.int(wrapLimit)

	//int SnapShotStructRun(SnapShotTask* sst, char* taskId, char* inputUrl, char* outputFolder, int timeInterval, int wrapLimit);
	ret := C.SegmentStructRun(taskIdC, inputUrlC, outputFolderC, timeInterval, wrapLimit)

	Info("task %s ret %d", cs.taskId, ret)
}

func (cs *CSegment) Stop() {
	taskIdC := C.CString(cs.taskId)
	defer C.free(unsafe.Pointer(taskIdC))

	C.StopTaskForGo(taskIdC)
}

// 关闭，释放资源
func (cs *CSegment) close() {

}
