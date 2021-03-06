#define _CRT_NONSTDC_NO_DEPRECATE
#define _CRT_SECURE_NO_WARNINGS

#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libswscale/swscale.h>
//#include <libavutil/timestamp.h>
#include <log.h>
#include <callback.h>

typedef enum TASK_STATE {
	RUNNING,
	STOP,
	FAULT,
} TASK_STATE;

// 内部结构体，不给外部使用
typedef struct Segment {

	char taskId[128];  // 任务id，给上层回调使用

	// 一些命名文件用的东西
	char nameBuffer[128];
	char inputUrl[128];
	char outputFolder[128];

	// ts的时间间隔 与 数量wrap
	int tsTimeInterval;       // 间隔多少时间进行一次切片，可以通过输入参数来自己定义，但是下限是GOP的大小，设置比GOP小就等于GOP
	int tsWrapLimit;          // 轮转数量，比如设置为 5，那么就会截图名称 0-4 轮转覆盖。-1 代表不轮转

	double tsBeginTime;       // ts起始时间
	double tsLastTime;        // ts结束时间

	int tsCount;
	int tsWrapCount;

	// 截图的时间间隔
	int snapTimeInterval;
	int snapWrapLimit;

	int snapCount;
	int snapWrapCount;

	AVOutputFormat* ofmt;
	AVFormatContext* ifmt_ctx, * ofmt_ctx;
	AVPacket* pkt;

	int video_index;
	int* stream_mapping;
	int stream_mapping_size;
} Segment;

int SegmentStructRun(char* taskId, char* inputUrl, char* outputFolder, int tsTimeInterval, int tsWrapLimit, int snapTimeInterval, int snapWrapLimit);

void StopTaskForGo(char* taskId);
