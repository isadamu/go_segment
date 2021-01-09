#include "_cgo_export.h"
#include <string.h>
#include <stdio.h>
#include <stdarg.h>
#include <log.h>
#include <libavcodec/avcodec.h>

#define MSG_MAX_LEN 512

enum FF_LOG_LEVEL ff_log_level = FF_INFO;

void GoDebug(const char* file, const char* func, int line, char* fmt, ...) {
	int i, j;
	char buf[MSG_MAX_LEN];
	i = snprintf(buf, MSG_MAX_LEN, "[%s][%s:%d] ", file, func, line);
	va_list arglist;
	va_start(arglist, fmt);
	j = vsnprintf(&buf[i], MSG_MAX_LEN - i, fmt, arglist);
	va_end(arglist);

	GoString go_str = { p: buf, n : i + j };
	DebugForC(go_str);
}


void GoInfo(const char* file, const char* func, int line, char* fmt, ...) {
	int i, j;
	char buf[MSG_MAX_LEN];
	i = snprintf(buf, MSG_MAX_LEN, "[%s][%s:%d] ", file, func, line);
	va_list arglist;
	va_start(arglist, fmt);
	j = vsnprintf(&buf[i], MSG_MAX_LEN - i, fmt, arglist);
	va_end(arglist);

	GoString go_str = { p: buf, n : i + j };
	InfoForC(go_str);
}

void GoWarn(const char* file, const char* func, int line, char* fmt, ...) {
	int i, j;
	char buf[MSG_MAX_LEN];
	i = snprintf(buf, MSG_MAX_LEN, "[%s][%s:%d] ", file, func, line);
	va_list arglist;
	va_start(arglist, fmt);
	j = vsnprintf(&buf[i], MSG_MAX_LEN - i, fmt, arglist);
	va_end(arglist);

	GoString go_str = { p: buf, n : i + j };
	WarnForC(go_str);
}

void GoError(const char* file, const char* func, int line, char* fmt, ...) {
	int i, j;
	char buf[MSG_MAX_LEN];
	i = snprintf(buf, MSG_MAX_LEN, "[%s][%s:%d] ", file, func, line);
	va_list arglist;
	va_start(arglist, fmt);
	j = vsnprintf(&buf[i], MSG_MAX_LEN - i, fmt, arglist);
	va_end(arglist);

	GoString go_str = { p: buf, n : i + j };
	ErrorForC(go_str);
}

void SetFFLogLevel(enum FF_LOG_LEVEL level) {
	ff_log_level = level;
}

void LogForFF(void* ptr, int level, const char* fmt, va_list vl) {
	if (level >= AV_LOG_DEBUG) {
		if (ff_log_level >= FF_DEBUG) {
			LogDebug(fmt, vl);
		}
	}
	else if (level >= AV_LOG_INFO) {
		if (ff_log_level >= FF_INFO) {
			LogInfo(fmt, vl);
		}
	}
	else if (level >= AV_LOG_WARNING) {
		if (ff_log_level >= FF_WARN) {
			LogWarn(fmt, vl);
		}
	}
	else if (level >= AV_LOG_QUIET) {
		if (ff_log_level >= FF_ERROR) {
			LogError(fmt, vl);
		}
	}
	else {
		LogError(fmt, vl);
	}
}