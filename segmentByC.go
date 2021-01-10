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
	"unsafe"
)

/**
ts切割，直接与底层c对接
*/

type CSegment struct {
	hasSentStop bool // 是否发送过结束命令

	config *SegmentEngineConfig
}

// 初始化
func NewCSegment(config *SegmentEngineConfig) *CSegment {

	cs := &CSegment{
		config: config,
	}

	return cs
}

// 开始截图
func (cs *CSegment) Run() int {
	taskIdC := C.CString(cs.config.taskId)
	inputUrlC := C.CString(cs.config.inputUrl)
	outputFolderC := C.CString(cs.config.outputFolder)

	defer C.free(unsafe.Pointer(taskIdC))
	defer C.free(unsafe.Pointer(inputUrlC))
	defer C.free(unsafe.Pointer(outputFolderC))

	tsTimeInterval := C.int(cs.config.tsTimeInterval)
	tsWrapLimit := C.int(cs.config.tsWrapLimit)
	snapTimeInterval := C.int(cs.config.snapTimeInterval)
	snapWrapLimit := C.int(cs.config.snapWrapLimit)

	// int SegmentStructRun(char* taskId, char* inputUrl, char* outputFolder, int tsTimeInterval, int tsWrapLimit, int snapTimeInterval, int snapWrapLimit);
	ret := C.SegmentStructRun(taskIdC, inputUrlC, outputFolderC, tsTimeInterval, tsWrapLimit, snapTimeInterval, snapWrapLimit)

	if ret < 0 {
		Error("task %s end, ret [%d]", cs.config.taskId, ret)
	} else {
		Info("task %s end, ret [%d]", cs.config.taskId, ret)
	}

	return ret
}

func (cs *CSegment) Stop() {
	taskIdC := C.CString(cs.config.taskId)
	defer C.free(unsafe.Pointer(taskIdC))

	C.StopTaskForGo(taskIdC)

	cs.hasSentStop = true
}

// 关闭，释放资源
// 这里需要先stop再往后
func (cs *CSegment) close() {
	if !cs.hasSentStop {
		cs.Stop()
	}
}
