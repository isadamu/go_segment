package main

import "C"

import (
	"fmt"
	"github.com/sasha-s/go-deadlock"
	"strconv"
)

//export SegmentCallBackForC
func SegmentCallBackForC(taskId, filePath string) {
	// 用于c语言回调，貌似只能这么猥琐的来写了

	fmt.Printf("taskId [%s], filePath [%s]\n", taskId, filePath)
	Info("taskId [%s], filePath [%s]", taskId, filePath)

}

const outputFolderHead = "segments/"

type SegmentManager struct {
	lock    deadlock.Mutex
	engines map[string]*SegmentEngine
}

func NewSegmentManager() *SegmentManager {
	return &SegmentManager{
		engines: make(map[string]*SegmentEngine),
	}
}

func (sm *SegmentManager) AddTask(vhost, app, streamName string, uid int, inputUrl string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	taskId := sm.genTaskId(vhost, app, streamName, uid)
	if _, ok := sm.engines[taskId]; ok {
		Info("task %s is already running", taskId)
		return
	}

	outputFolder := outputFolderHead + vhost + "-" + app + "-" + streamName + "-" + strconv.Itoa(uid)

	engine := NewSegmentEngine(taskId, inputUrl, outputFolder)

	engine.Start()

	sm.engines[taskId] = engine
}

func (sm *SegmentManager) DelTask(vhost, app, streamName string, uid int) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	taskId := sm.genTaskId(vhost, app, streamName, uid)

	engine, ok := sm.engines[taskId]
	if !ok {
		Info("task %s is not exist", taskId)
		return
	}

	engine.Stop()
	engine.close()

	delete(sm.engines, taskId)

}

func (sm *SegmentManager) genTaskId(vhost, app, streamName string, uid int) string {
	return vhost + "/" + app + "/" + streamName + "/" + strconv.Itoa(uid)
}
