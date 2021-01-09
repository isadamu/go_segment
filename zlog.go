package main

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"runtime"
	"strings"
)

const (
	LevelDebug = iota //0
	LevelInfo
	LevelWarn
	LevelError
)

var programInfo string

var loglevel int
var logger *zap.SugaredLogger

func MyCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	str := programInfo + "[" + caller.TrimmedPath() + "][" + runtime.FuncForPC(caller.PC).Name() + "]"
	enc.AppendString(str)
}

func Init(filePath string, maxSize, maxAge, maxBackups int, compress bool, level int) {
	genFmtStr()

	loglevel = level

	logFile := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,    // megabytes
		MaxAge:     maxAge,     // 0 表示不限制
		MaxBackups: maxBackups, // 0 表示不限制
		Compress:   compress,   // disabled by default
		LocalTime:  true,
	}
	w := zapcore.AddSync(logFile)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeCaller = MyCallerEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		zapcore.DebugLevel,
	)
	caller := zap.AddCaller()
	callerSkip := zap.AddCallerSkip(1)
	zapLogger := zap.New(core, caller, callerSkip)
	logger = zapLogger.Sugar()
}

func genFmtStr() {
	procName := os.Args[0]
	idx := strings.LastIndex(procName, "/")
	programName := procName[idx+1:]

	programPid := os.Getpid()

	programInfo = fmt.Sprintf("[%d][%s]", programPid, programName)
}

func SetLevel(level int) {
	loglevel = level
}

func GetLevel() int {
	return loglevel
}

func LogLevelToString(level int) string {
	switch level {
	case LevelError:
		{
			return "error"
		}
	case LevelWarn:
		{
			return "warn"
		}
	case LevelInfo:
		{
			return "info"
		}
	case LevelDebug:
		{
			return "debug"
		}
	default:
		return "unknown level"
	}
}

func Debug(format string, args ...interface{}) {
	if loglevel > LevelDebug {
		return
	}

	logger.Debugf(format, args...)
}

func Info(format string, args ...interface{}) {
	if loglevel > LevelInfo {
		return
	}

	logger.Infof(format, args...)
}

func Warn(format string, args ...interface{}) {
	if loglevel > LevelWarn {
		return
	}

	logger.Warnf(format, args...)
}

func Error(format string, args ...interface{}) {
	if loglevel > LevelError {
		return
	}

	logger.Errorf(format, args...)
}
