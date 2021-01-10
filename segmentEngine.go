package main

import (
	"errors"
	"github.com/sasha-s/go-deadlock"
	"os"
	"strings"
)

/**
管理ts切割的中间层
*/

var (
	ErrEngineIsNotInReady             = errors.New("engine is not in ready state")
	ErrEngineCreateOutputFolderFailed = errors.New("engine create output folder failed")
)

type SegmentEngineState int

const (
	SegmentEngineStateReady   SegmentEngineState = iota + 1 // 可以运行状态
	SegmentEngineStateRunning                               // 运行状态
	SegmentEngineStateStop                                  // 正常停止状态
	SegmentEngineStateError                                 // 出错，被迫停止状态
	SegmentEngineStateClose                                 // 关闭
)

func (state SegmentEngineState) String() string {
	switch state {
	case SegmentEngineStateReady:
		return "ready"
	case SegmentEngineStateRunning:
		return "running"
	case SegmentEngineStateStop:
		return "stop"
	case SegmentEngineStateError:
		return "error"
	case SegmentEngineStateClose:
		return "close"
	default:
		return "unknown"
	}
}

type SegmentEngineConfig struct {
	taskId           string
	inputUrl         string
	outputFolder     string
	tsTimeInterval   int
	tsWrapLimit      int
	snapTimeInterval int
	snapWrapLimit    int
}

type SegmentEngine struct {
	sgMgr *SegmentManager

	lock  deadlock.Mutex
	state SegmentEngineState

	config *SegmentEngineConfig

	core *CSegment // 使用包装好的c来进行截图
}

// 初始化
func NewSegmentEngine(sgMgr *SegmentManager, config SegmentEngineConfig) (*SegmentEngine, error) {

	pConfig := &config

	if strings.HasSuffix(pConfig.outputFolder, "/") {
		pConfig.outputFolder = pConfig.outputFolder[:len(pConfig.outputFolder)-1]
	}

	err := os.MkdirAll(pConfig.outputFolder, 0755)
	if err != nil {
		Error("task %s create output dir %s err: %s\n", pConfig.taskId, pConfig.outputFolder, err)
		return nil, ErrEngineCreateOutputFolderFailed
	}

	segmentCore := NewCSegment(pConfig)

	return &SegmentEngine{
		sgMgr:  sgMgr,
		state:  SegmentEngineStateReady,
		config: pConfig,
		core:   segmentCore,
	}, nil
}

// 开始截图
func (se *SegmentEngine) Start() error {
	if se.GetState() != SegmentEngineStateReady {
		return ErrEngineIsNotInReady
	}

	go se.runTheEngine() // 需要新开go程，不然会阻塞

	return nil
}

// 这个函数在底层运行时会阻塞，如果底层被停止或者意外退出时才会解除
// 停止时通知上层进行处理
func (se *SegmentEngine) runTheEngine() {
	se.setState(SegmentEngineStateRunning)

	ret := se.core.Run() // 这里会阻塞

	Info("engine %s stop, ret %d", se.config.taskId, ret)

}

// 停止截图
func (se *SegmentEngine) Stop() {
	se.core.Stop()
}

func (se *SegmentEngine) IsRunning() bool {
	se.lock.Lock()
	defer se.lock.Unlock()

	return se.state == SegmentEngineStateRunning
}

func (se *SegmentEngine) setState(state SegmentEngineState) {
	se.lock.Lock()
	defer se.lock.Unlock()

	se.state = state
}

func (se *SegmentEngine) GetState() SegmentEngineState {
	se.lock.Lock()
	defer se.lock.Unlock()

	return se.state
}

// 关闭，释放资源
func (se *SegmentEngine) close() {

	if se.GetState() == SegmentEngineStateRunning {
		se.core.Stop()
	}

	se.setState(SegmentEngineStateClose)

	if se.core != nil {
		se.core.close()
		se.core = nil
	}
}
