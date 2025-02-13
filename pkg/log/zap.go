package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logTmFmt = "2006-01-02 15:04:05"
)

func NewLogger(file string, debug bool) *zap.Logger {
	level := zapcore.DebugLevel
	if !debug {
		level = zapcore.InfoLevel
	}
	// 文件日志: 不显示Debug级别日志
	// 控制台日志: 彩色显示
	newCore := zapcore.NewTee(
		zapcore.NewCore(GetEncoder(cEncodeLevel), GetWriteSyncer(file), zapcore.InfoLevel),            // 写入文件
		zapcore.NewCore(GetEncoder(zapcore.CapitalColorLevelEncoder), zapcore.Lock(os.Stdout), level), // 写入控制台
	)
	return zap.New(newCore, zap.AddCaller(), zap.AddCallerSkip(1))

}

// GetEncoder 自定义的Encoder
func GetEncoder(levelEncoder zapcore.LevelEncoder) zapcore.Encoder {
	conf := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller_line",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    levelEncoder,
		EncodeTime:     cEncodeTime,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   cEncodeCaller,
	}
	return zapcore.NewConsoleEncoder(conf)
}

// GetWriteSyncer 自定义的WriteSyncer
func GetWriteSyncer(file string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    8,
		MaxBackups: 10,
		MaxAge:     14,
		LocalTime:  true,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// cEncodeLevel 自定义日志级别显示
func cEncodeLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

// cEncodeTime 自定义时间格式显示
func cEncodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + t.Format(logTmFmt) + "]")
}

// cEncodeCaller 自定义行号显示
func cEncodeCaller(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + caller.TrimmedPath() + "]")
}
