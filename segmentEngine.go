package main

import (
	"github.com/sasha-s/go-deadlock"
	"os"
)

type TaskState int

const (
	TaskStateRunning TaskState = iota
	TaskStateStop
)

func (ts TaskState) String() string {
	switch ts {
	case TaskStateRunning:
		return "running"
	case TaskStateStop:
		return "stop"
	default:
		return "unknown"
	}
}

type SegmentEngine struct {
	lock deadlock.Mutex

	taskId string // vhost/app/streamName

	taskState TaskState

	inputUrl   string // 输入的URL，应该是一个rtmp拉流链接
	outputPath string // TS文件临时存放的位置

	core *CSegment // 使用包装好的c来进行截图
}

// 初始化
func NewSegmentEngine(taskId, inputUrl, outputFolder string) *SegmentEngine {

	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		Error("Create snapshot path %s err: %s\n", outputFolder, err)
		return nil
	}

	segmentCore := NewCSegment(taskId, inputUrl, outputFolder)

	return &SegmentEngine{
		taskId:     taskId,
		inputUrl:   inputUrl,
		outputPath: outputFolder,
		core:       segmentCore,
	}
}

// 开始截图
func (se *SegmentEngine) Start() {
	go se.core.Start() // 需要新开go程，不然会阻塞

	se.lock.Lock()
	defer se.lock.Unlock()
	se.taskState = TaskStateRunning
}

// 停止截图
func (se *SegmentEngine) Stop() {
	se.lock.Lock()
	defer se.lock.Unlock()

	se.core.Stop()

	se.taskState = TaskStateStop
}

func (se *SegmentEngine) IsRunning() bool {
	se.lock.Lock()
	defer se.lock.Unlock()

	return se.taskState == TaskStateRunning
}

// 关闭，释放资源
func (se *SegmentEngine) close() {
	se.lock.Lock()
	defer se.lock.Unlock()

	se.taskState = TaskStateStop

	if se.core != nil {
		se.core.close()
		se.core = nil
	}
}
