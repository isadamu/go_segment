#define _CRT_NONSTDC_NO_DEPRECATE
#define _CRT_SECURE_NO_WARNINGS

#define LogDebug(fmt, ...) GoDebug(__FILE__, __FUNCTION__, __LINE__, fmt, ##__VA_ARGS__)
#define LogInfo(fmt, ...) GoInfo(__FILE__, __FUNCTION__, __LINE__, fmt, ##__VA_ARGS__)
#define LogWarn(fmt, ...) GoWarn(__FILE__, __FUNCTION__, __LINE__, fmt, ##__VA_ARGS__)
#define LogError(fmt, ...) GoError(__FILE__, __FUNCTION__, __LINE__, fmt, ##__VA_ARGS__)

typedef enum FF_LOG_LEVEL {
	FF_ERROR,
	FF_WARN,
	FF_INFO,
	FF_DEBUG,
} FF_LOG_LEVEL;

void GoDebug(const char* file, const char* func, int line, char* fmt, ...);
void GoInfo(const char* file, const char* func, int line, char* fmt, ...);
void GoWarn(const char* file, const char* func, int line, char* fmt, ...);
void GoError(const char* file, const char* func, int line, char* fmt, ...);

void SetFFLogLevel(enum FF_LOG_LEVEL level);
void LogForFF(void* ptr, int level, const char* fmt, va_list vl);
