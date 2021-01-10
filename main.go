package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

const (
	LogFilePath   = "test.log"
	LogMaxSize    = 200
	LogMaxAge     = 0
	LogMaxBackups = 0
)

const (
	outputFolderHead = "segments"
	tsTimeInterval   = 20
	tsWrapLimit      = 3
	snapTimeInterval = 20
	snapWrapLimit    = 3
)

var SegmentMgr *SegmentManager = nil

func main() {
	// 设置多核
	maxProces := runtime.NumCPU()
	if maxProces > 1 {
		maxProces -= 1
	}
	runtime.GOMAXPROCS(maxProces)

	Init(LogFilePath, LogMaxSize, LogMaxAge, LogMaxBackups, true, LevelInfo)

	SetLevel(LevelDebug)

	Info("begin")

	if len(os.Args) != 3 {
		_, _ = fmt.Fprintf(os.Stderr, "input params: <inputUrl> <num>\n")
		return
	}

	inputUrl := os.Args[1]

	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "input params err: %s", err)
		return
	}

	// 捕捉系统信号
	signalChan := getSignalChan()

	SegmentMgr = NewSegmentManager()
	SetCallBackMgr(SegmentMgr)

	vhost := "aaa"
	app := "bbb"
	streamName := "ccc"

	for i := 0; i < num; i++ {
		config := SegmentTaskConfig{
			vhost:            vhost,
			app:              app,
			streamName:       streamName,
			inputUrl:         inputUrl,
			outputFolderHead: outputFolderHead,
			tsTimeInterval:   tsTimeInterval,
			tsWrapLimit:      tsWrapLimit,
			snapTimeInterval: snapTimeInterval,
			snapWrapLimit:    snapWrapLimit,
		}
		err := SegmentMgr.AddTask(config)
		if err != nil {
			fmt.Printf("addTask failed: %s\n", err)
		}
		time.Sleep(time.Millisecond * 200)
	}

	//使用C.CString创建的字符串需要手动释放。
	////////////////////////////////////////////
	//inputUrl0 := "rtmp://61.129.131.5/live/lrm0-"
	//urlSuffix0 := "?vhost=lrm.pull-dev.kijazz.cn"
	//vhost0 := "lrm.pull-dev.kijazz.cn"
	//app0 := "live"
	//streamName0 := "lrm0-"
	//for i := 0; i < num; i++ {
	//	inputUrlWithNum := inputUrl0 + strconv.Itoa(i) + urlSuffix0
	//	streamNameWithNum := streamName0 + strconv.Itoa(i)
	//	SegmentMgr.AddTask(vhost0, app0, streamNameWithNum, i, inputUrlWithNum)
	//	time.Sleep(time.Millisecond * 20)
	//}
	//
	/////////////////////////////////////////////
	//inputUrl1 := "rtmp://61.129.131.5/live/lrm1-"
	//urlSuffix1 := "?vhost=lrm.pull-dev.kijazz.cn"
	//vhost1 := "lrm.pull-dev.kijazz.cn"
	//app1 := "live"
	//streamName1 := "lrm1-"
	//for i := 0; i < num; i++ {
	//	inputUrlWithNum := inputUrl1 + strconv.Itoa(i) + urlSuffix1
	//	streamNameWithNum := streamName1 + strconv.Itoa(i)
	//	SegmentMgr.AddTask(vhost1, app1, streamNameWithNum, i, inputUrlWithNum)
	//	time.Sleep(time.Millisecond * 20)
	//}
	//
	/////////////////////////////////////////////
	//inputUrl2 := "rtmp://61.129.131.5/live/lrm2-"
	//urlSuffix2 := "?vhost=lrm.pull-dev.kijazz.cn"
	//vhost2 := "lrm.pull-dev.kijazz.cn"
	//app2 := "live"
	//streamName2 := "lrm2-"
	//for i := 0; i < num; i++ {
	//	inputUrlWithNum := inputUrl2 + strconv.Itoa(i) + urlSuffix2
	//	streamNameWithNum := streamName2 + strconv.Itoa(i)
	//	SegmentMgr.AddTask(vhost2, app2, streamNameWithNum, i, inputUrlWithNum)
	//	time.Sleep(time.Millisecond * 20)
	//}

	sig := <-signalChan
	Info("receive kill signal %s", sig)

	for i := 0; i < num; i++ {
		SegmentMgr.DelTask(vhost, app, streamName, i)
		time.Sleep(time.Millisecond * 20)
	}

	//for i := 0; i < num; i++ {
	//	streamNameWithNum := streamName0 + strconv.Itoa(i)
	//	SegmentMgr.DelTask(vhost0, app0, streamNameWithNum, i)
	//	time.Sleep(time.Millisecond * 10)
	//}
	//
	//for i := 0; i < num; i++ {
	//	streamNameWithNum := streamName1 + strconv.Itoa(i)
	//	SegmentMgr.DelTask(vhost1, app1, streamNameWithNum, i)
	//	time.Sleep(time.Millisecond * 10)
	//}
	//
	//for i := 0; i < num; i++ {
	//	streamNameWithNum := streamName2 + strconv.Itoa(i)
	//	SegmentMgr.DelTask(vhost2, app2, streamNameWithNum, i)
	//	time.Sleep(time.Millisecond * 10)
	//}

	time.Sleep(time.Second * 1)

	Info("bye")

}

func getSignalChan() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)
	signal.Notify(signalChan, syscall.SIGTERM)

	return signalChan
}
