// 参考 https://www.cnblogs.com/leisure_chn/p/10584901.html
// https://blog.csdn.net/lightfish_zhang/article/details/86594694

#define _CRT_NONSTDC_NO_DEPRECATE
#define _CRT_SECURE_NO_WARNINGS

#include <segment.h>
#include <map.h>
#include <pthread.h>

pthread_mutex_t taskMapMute;
int mapInit = 0;
map_int_t taskStateMap;

pthread_mutex_t ffLogInitMute;
int ffLogInit = 0;

void initFFLogCallBack() {
	pthread_mutex_lock(&ffLogInitMute);
	if (!ffLogInit) {
		ffLogInit = 1;
		av_log_set_callback(LogForFF);
	}
	pthread_mutex_unlock(&ffLogInitMute);
}

void initTaskStateMap() {
	pthread_mutex_lock(&taskMapMute);
	if (!mapInit) {
		mapInit = 1;
		map_init(&taskStateMap);
	}
	pthread_mutex_unlock(&taskMapMute);
}

// 全局初始化
void initGlobal() {
	//initFFLogCallBack();   // 这个会将ffmpeg的日志传送到go中，可以不开启
	initTaskStateMap();      // 初始化任务的状态队列
}

void setTaskState(char* taskId, int state) {
	pthread_mutex_lock(&taskMapMute);
	map_set(&taskStateMap, taskId, state);
	pthread_mutex_unlock(&taskMapMute);
}

void delTaskState(char* taskId) {
	pthread_mutex_lock(&taskMapMute);
	map_remove(&taskStateMap, taskId);
	pthread_mutex_unlock(&taskMapMute);
}

int* getTaskState(char* taskId) {
	pthread_mutex_lock(&taskMapMute);
	int* val = map_get(&taskStateMap, taskId);
	pthread_mutex_unlock(&taskMapMute);
	return val;
}

int isTaskStateRunning(char* taskId) {
	int* state = getTaskState(taskId);
	if (state == NULL) {
		return 0;
	}
	return *state == RUNNING;
}

int interruptCallBack(void* taskId) {
	if (isTaskStateRunning((char*)taskId)) {
		return 0;
	}
	return 1;
}

void StopTaskForGo(char* taskId) {
	setTaskState(taskId, STOP);
}

static void log_packet(const AVFormatContext* fmt_ctx, const AVPacket* pkt, const char* tag)
{
	AVRational* time_base = &fmt_ctx->streams[pkt->stream_index]->time_base;

	printf("%s: pts:%s pts_time:%s dts:%s dts_time:%s duration:%s duration_time:%s stream_index:%d\n",
		tag,
		av_ts2str(pkt->pts), av_ts2timestr(pkt->pts, time_base),
		av_ts2str(pkt->dts), av_ts2timestr(pkt->dts, time_base),
		av_ts2str(pkt->duration), av_ts2timestr(pkt->duration, time_base),
		pkt->stream_index);
}


// 初始化解码相关变量
int openInput(Segment* ss) {
	int ret;

	// 先申请输入的context，添加 interruptCallBack
	// 作用是当输入流停止时，仍能通过这个 callback 来控制是否跳出阻塞
	ss->ifmt_ctx = avformat_alloc_context();

	ss->ifmt_ctx->interrupt_callback.callback = interruptCallBack;
	ss->ifmt_ctx->interrupt_callback.opaque = ss->taskId;

	if ((ret = avformat_open_input(&(ss->ifmt_ctx), ss->inputUrl, 0, 0)) < 0) {
		LogError("task %s could not open input '%s'", ss->taskId, ss->inputUrl);
		return ret;
	}

	if ((ret = avformat_find_stream_info(ss->ifmt_ctx, 0)) < 0) {
		LogError("task %s failed to retrieve input stream information", ss->taskId);
		return ret;
	}

	av_dump_format(ss->ifmt_ctx, 0, ss->inputUrl, 0);

	ss->stream_mapping_size = ss->ifmt_ctx->nb_streams;
	ss->stream_mapping = av_mallocz_array(ss->stream_mapping_size, sizeof(*(ss->stream_mapping)));
	if (!ss->stream_mapping) {
		ret = AVERROR(ENOMEM);
		return ret;
	}

	int i, stream_index = 0;
	for (i = 0; i < ss->ifmt_ctx->nb_streams; i++) {
		AVStream* in_stream = ss->ifmt_ctx->streams[i];
		AVCodecParameters* in_codecpar = in_stream->codecpar;

		if (in_codecpar->codec_type != AVMEDIA_TYPE_AUDIO &&
			in_codecpar->codec_type != AVMEDIA_TYPE_VIDEO &&
			in_codecpar->codec_type != AVMEDIA_TYPE_SUBTITLE) {
			ss->stream_mapping[i] = -1;
			continue;
		}

		if (in_codecpar->codec_type == AVMEDIA_TYPE_VIDEO) {
			ss->video_index = i;
		}

		ss->stream_mapping[i] = stream_index++;
	}

	return 0;
}

int openOutput(Segment* ss) {
	int ret = 0;

	ss->tsCount++;

	freeOutput(ss);

	constuctFileName(ss);

	avformat_alloc_output_context2(&(ss->ofmt_ctx), NULL, NULL, ss->nameBuffer);
	if (!(ss->ofmt_ctx)) {
		LogError("task %s could not create output context", ss->taskId);
		ret = AVERROR_UNKNOWN;
		return ret;
	}

	ss->ofmt = ss->ofmt_ctx->oformat;

	int i;
	for (i = 0; i < ss->ifmt_ctx->nb_streams; i++) {

		if (ss->stream_mapping[i] == -1) {
			continue;
		}

		AVStream* out_stream;
		AVStream* in_stream = ss->ifmt_ctx->streams[i];
		AVCodecParameters* in_codecpar = in_stream->codecpar;

		out_stream = avformat_new_stream(ss->ofmt_ctx, NULL);
		if (!out_stream) {
			LogError("task %s failed allocating output stream", ss->taskId);
			ret = AVERROR_UNKNOWN;
			return ret;
		}

		ret = avcodec_parameters_copy(out_stream->codecpar, in_codecpar);
		if (ret < 0) {
			LogError("task %s failed to copy codec parameters", ss->taskId);
			return ret;
		}
		out_stream->codecpar->codec_tag = 0;
	}

	//av_dump_format(ss->ofmt_ctx, 0, ss->nameBuffer, 1);

	if (!(ss->ofmt->flags & AVFMT_NOFILE)) {
		ret = avio_open(&(ss->ofmt_ctx->pb), ss->nameBuffer, AVIO_FLAG_WRITE);
		if (ret < 0) {
			LogError("task %s could not open output file '%s'", ss->taskId, ss->nameBuffer);
			return ret;
		}
	}

	ret = avformat_write_header(ss->ofmt_ctx, NULL);
	if (ret < 0) {
		LogError("task %s error occurred when write header to output file '%s'", ss->taskId, ss->nameBuffer);
		return ret;
	}


	return 0;
}


void freeInput(Segment* ss) {
	av_packet_free(&(ss->pkt));
	avformat_free_context(ss->ifmt_ctx);
}

void freeOutput(Segment* ss) {
	if (ss->ofmt_ctx != NULL) {
		av_write_trailer(ss->ofmt_ctx);
		GoSegmentCallBackForC(ss->taskId, ss->nameBuffer);
	}
	if (ss->ofmt_ctx && !(ss->ofmt_ctx->flags & AVFMT_NOFILE)) {
		avio_closep(&(ss->ofmt_ctx->pb));
	}
	if (ss->ofmt_ctx != NULL) {
		avformat_free_context(ss->ofmt_ctx);
		ss->ofmt_ctx = NULL;
	}
}


// 构建输出文件名，例如 0.ts  1.ts 2.ts 前面假设文件夹路径
void constuctFileName(Segment* ss) {
	char intBuf[32];
	sprintf(intBuf, "%d", ss->wrap_count);

	memset(ss->nameBuffer, 0, sizeof(ss->nameBuffer));
	strcat(ss->nameBuffer, ss->outputFolder);
	strcat(ss->nameBuffer, "/");
	strcat(ss->nameBuffer, intBuf);
	strcat(ss->nameBuffer, ".ts");

	ss->wrap_count = (ss->wrap_count + 1) % ss->wrap_limit;
}

Segment* initSegmentStruct(char* taskId, char* inputUrl, char* outputFolder, int timeInterval, int wrapLimit) {

	Segment* ss = (Segment*)malloc(sizeof(Segment));

	strcpy(ss->taskId, taskId);
	strcpy(ss->inputUrl, inputUrl);
	strcpy(ss->outputFolder, outputFolder);

	ss->time_interval = timeInterval;
	ss->wrap_limit = wrapLimit;

	ss->tsCount = 0;
	ss->wrap_count = 0;

	ss->ifmt_ctx = NULL;
	ss->ofmt_ctx = NULL;
	ss->ofmt = NULL;
	ss->pkt = NULL;

	return ss;
}

int SegmentStructRun(char* taskId, char* inputUrl, char* outputFolder, int timeInterval, int wrapLimit) {

	// 初始化各种参数
	initGlobal();

	int ret;

	Segment* ss = initSegmentStruct(taskId, inputUrl, outputFolder, timeInterval, wrapLimit);

	// 设置任务状态
	setTaskState(ss->taskId, RUNNING);

	LogInfo("------------- task %s begin -------------", ss->taskId);

	// openInput
	ret = openInput(ss);
	if (ret < 0) {
		goto end;
	}

	LogInfo("------------- task %s open input '%s' success -------------", ss->taskId, ss->inputUrl);


	/**********************************************************************/
	/******************************** 分配pkt *****************************/
	ss->pkt = av_packet_alloc();
	if (!ss->pkt) {
		LogError("task %s could not allocate pkt", ss->taskId);
		ret = AVERROR_UNKNOWN;
		goto end;
	}

	/**********************************************************************/
	/****************************** 循环 切 ts ***************************/
	int isFirst = 1;
	double beginTime = -1, lastTime = -1;
	while (1) {

		ret = av_read_frame(ss->ifmt_ctx, ss->pkt);
		if (ret < 0) {
			break;
		}

		// 判断是不是已经被上层通知结束了
		if (!isTaskStateRunning(ss->taskId)) {
			LogWarn("------------- task %s state is stopped, break the loop -------------", ss->taskId);
			break;
		}

		// 首先一定要创建文件
		if (isFirst) {
			isFirst = 0;

			double timeStamp = ss->pkt->pts * av_q2d(ss->ifmt_ctx->streams[ss->video_index]->time_base);
			beginTime = timeStamp;
			lastTime = timeStamp;
			ret = openOutput(ss);
			if (ret < 0) {
				goto end;
			}
		}

		// 遇到关键帧判断一下是否要新开切片
		if ((ss->pkt->stream_index == ss->video_index) && (ss->pkt->flags & AV_PKT_FLAG_KEY)) {
			double timeStamp = ss->pkt->pts * av_q2d(ss->ifmt_ctx->streams[ss->video_index]->time_base);

			lastTime = timeStamp;
			if ((lastTime - beginTime) >= ss->time_interval) {
				beginTime = timeStamp;
				ret = openOutput(ss);
				if (ret < 0) {
					goto end;
				}
			}
		}

		AVStream* in_stream, * out_stream;

		in_stream = ss->ifmt_ctx->streams[ss->pkt->stream_index];
		if (ss->pkt->stream_index >= ss->stream_mapping_size ||
			ss->stream_mapping[ss->pkt->stream_index] < 0) {
			av_packet_unref(ss->pkt);
			continue;
		}

		ss->pkt->stream_index = ss->stream_mapping[ss->pkt->stream_index];
		out_stream = ss->ofmt_ctx->streams[ss->pkt->stream_index];

		// 调试使用
		log_packet(ss->ifmt_ctx, ss->pkt, "in");

		/* copy packet */
		ss->pkt->pts = av_rescale_q_rnd(ss->pkt->pts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF | AV_ROUND_PASS_MINMAX);
		ss->pkt->dts = av_rescale_q_rnd(ss->pkt->dts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF | AV_ROUND_PASS_MINMAX);
		ss->pkt->duration = av_rescale_q(ss->pkt->duration, in_stream->time_base, out_stream->time_base);
		ss->pkt->pos = -1;

		// 调试使用
		log_packet(ss->ofmt_ctx, ss->pkt, "out");

		ret = av_interleaved_write_frame(ss->ofmt_ctx, ss->pkt);
		if (ret < 0) {
			if (ret == -22) {

			}
			LogInfo("task %s error muxing packet", ss->taskId);
			break;
		}
		av_packet_unref(ss->pkt);
	}

end:

	if (ret < 0) {
		LogError("task %s ret %d error occurred: %s", ss->taskId, ret, av_err2str(ret));
	}
	else {
		LogInfo("task %s exit success", ss->taskId);
	}

	// 释放资源
	freeOutput(ss);

	freeInput(ss);

	delTaskState(ss->taskId);

	free(ss);

	return ret;
}
