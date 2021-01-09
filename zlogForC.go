package main

import "C"

//export DebugForC
func DebugForC(message string) {
	if loglevel > LevelDebug {
		return
	}

	logger.Debug(message)
}

//export InfoForC
func InfoForC(message string) {
	if loglevel > LevelInfo {
		return
	}

	logger.Info(message)
}

//export WarnForC
func WarnForC(message string) {
	if loglevel > LevelWarn {
		return
	}

	logger.Warnf(message)
}

//export ErrorForC
func ErrorForC(message string) {
	if loglevel > LevelError {
		return
	}

	logger.Errorf(message)
}
