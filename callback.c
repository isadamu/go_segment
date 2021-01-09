#include "_cgo_export.h"
#include <string.h>

void GoSegmentCallBackForC(char* taskId, char* filePath) {
	GoString go_taskId = {p: taskId, n: strlen(taskId)};
	GoString go_filePath = {p: filePath, n: strlen(filePath)};

	SegmentCallBackForC(go_taskId, go_filePath);
}