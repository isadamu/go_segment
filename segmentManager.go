package main

import "C"

import (
	"errors"
	"github.com/sasha-s/go-deadlock"
	"strconv"
)

var (
	ErrTaskIsAlreadyExist = errors.New("task is already exist")
)

type SegmentTaskConfig struct {
	vhost            string
	app              string
	streamName       string
	inputUrl         string
	outputFolderHead string

	tsTimeInterval   int
	tsWrapLimit      int
	snapTimeInterval int
	snapWrapLimit    int
}

type SegmentManager struct {
	lock    deadlock.Mutex
	engines map[string]*SegmentEngine

	uidLock deadlock.Mutex
	uid     int

	outputFolderHead string
}

func NewSegmentManager() *SegmentManager {
	return &SegmentManager{
		engines: make(map[string]*SegmentEngine),
	}
}

func (sm *SegmentManager) AddTask(taskConfig SegmentTaskConfig) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	uid := sm.genUid()
	taskId := sm.genTaskId(taskConfig.vhost, taskConfig.app, taskConfig.streamName, uid)
	if _, ok := sm.engines[taskId]; ok { // 这个不可能发生，先就放这里吧
		Error("task %s is already exists", taskId)
		return ErrTaskIsAlreadyExist
	}

	outputFolder := sm.genOutputFolder(taskConfig.vhost, taskConfig.app, taskConfig.streamName, uid)

	engineConfig := SegmentEngineConfig{
		taskId:           taskId,
		inputUrl:         taskConfig.inputUrl,
		outputFolder:     outputFolder,
		tsTimeInterval:   taskConfig.tsTimeInterval,
		tsWrapLimit:      taskConfig.tsWrapLimit,
		snapTimeInterval: taskConfig.snapTimeInterval,
		snapWrapLimit:    taskConfig.snapWrapLimit,
	}

	engine, err := NewSegmentEngine(sm, engineConfig)
	if err != nil {
		Error("create segment engine %s failed: %s", taskId, err)
		return err
	}

	err = engine.Start()
	if err != nil {
		Error("start segment engine %s err: %s", taskId, err)
		engine.close()
		return err
	}

	sm.engines[taskId] = engine

	return nil
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

func (sm *SegmentManager) genOutputFolder(vhost, app, streamName string, uid int) string {
	return sm.outputFolderHead + "/" + vhost + "/" + app + "/" + streamName + "/" + strconv.Itoa(uid)
}

func (sm *SegmentManager) genUid() int {
	sm.uidLock.Lock()
	defer sm.uidLock.Unlock()

	uid := sm.uid
	sm.uid++

	return uid
}
