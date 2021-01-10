#include "_cgo_export.h"
#include <string.h>

void GoTsCallBackForC(char* taskId, char* filePath, double tsBegineTime, double tsEndTime) {
	GoString go_taskId = {p: taskId, n: strlen(taskId)};
	GoString go_filePath = {p: filePath, n: strlen(filePath)};

	TsCallBackForC(go_taskId, go_filePath, tsBeginTime, tsEndTime);
}

void GoSnapCallBackForC(char* taskId, char* filePath, double snapTime) {
	GoString go_taskId = { p: taskId, n : strlen(taskId) };
	GoString go_filePath = { p: filePath, n : strlen(filePath) };

	SnapCallBackForC(go_taskId, go_filePath, snapTime);
}