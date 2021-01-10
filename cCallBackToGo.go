package main

import "fmt"

/**
这个函数用于c底层回调使用
*/

var segmentMgr *SegmentManager = nil

func SetCallBackMgr(sgMgr *SegmentManager) {
	segmentMgr = sgMgr
}

//export TsCallBackForC
func TsCallBackForC(taskId, filePath string, tsBeginTime, tsEndTime float64) {
	// 用于c语言回调，貌似只能这么猥琐的来写了，比较简单

	fmt.Printf("taskId [%s], filePath [%s], tsBeginTime [%f], tsEndTime [%f]\n", taskId, filePath, tsBeginTime, tsEndTime)
	Info("taskId [%s], filePath [%s], tsBeginTime [%f], tsEndTime [%f]", taskId, filePath, tsBeginTime, tsEndTime)

	if segmentMgr != nil {

	}

}

//export SnapCallBackForC
func SnapCallBackForC(taskId, filePath string, snapTime float64) {
	// 用于c语言回调，貌似只能这么猥琐的来写了，比较简单

	fmt.Printf("taskId [%s], filePath [%s]\n", taskId, filePath)
	Info("taskId [%s], filePath [%s]", taskId, filePath)

	if segmentMgr != nil {

	}

}
