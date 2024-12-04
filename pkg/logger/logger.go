package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log   *zap.Logger
	sugar *zap.SugaredLogger
)

// Init 初始化日志
func Init(level string, format string, output string) error {
	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	}

	// 控制台编码器配置（带颜色）
	consoleEncoderConfig := encoderConfig
	consoleEncoderConfig.EncodeLevel = CustomLevelEncoder
	consoleEncoderConfig.EncodeTime = TimeEncoder
	consoleEncoderConfig.EncodeCaller = CustomCallerEncoder

	// 创建编码器
	var consoleEncoder, fileEncoder zapcore.Encoder
	if format == "json" {
		consoleEncoder = zapcore.NewJSONEncoder(consoleEncoderConfig)
		fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
		fileEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置输出
	var writers []zapcore.WriteSyncer
	writers = append(writers, zapcore.AddSync(os.Stdout)) // 始终输出到控制台

	if output == "file" || output == "both" {
		// 确保日志目录存在
		if err := os.MkdirAll("logs", 0755); err != nil {
			return err
		}
		file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		_ = append(writers, zapcore.AddSync(file))
	}

	// 配置日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 创建核心
	var cores []zapcore.Core

	// 添加控制台输出
	if output == "stdout" || output == "both" {
		cores = append(cores, zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			zapLevel,
		))
	}

	// 添加文件输出
	if output == "file" || output == "both" {
		// 确保日志目录存在
		if err := os.MkdirAll("logs", 0755); err != nil {
			return err
		}
		file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		cores = append(cores, zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(file),
			zapLevel,
		))
	}

	// 创建日志记录器
	core := zapcore.NewTee(cores...)
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = log.Sugar()
	return nil
}

// TimeEncoder 自定义时间编码器
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("\x1b[90m" + t.Format("2006-01-02 15:04:05.000") + "\x1b[0m")
}

// CustomCallerEncoder 自定义调用位置编码器
func CustomCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("\x1b[90m" + caller.TrimmedPath() + "\x1b[0m")
}

// CustomLevelEncoder 自定义日志级别的颜色输出
func CustomLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var level string
	switch l {
	case zapcore.DebugLevel:
		level = "\x1b[36mDEBUG\x1b[0m" // 青色
	case zapcore.InfoLevel:
		level = "\x1b[32mINFO\x1b[0m" // 绿色
	case zapcore.WarnLevel:
		level = "\x1b[33mWARN\x1b[0m" // 黄色
	case zapcore.ErrorLevel:
		level = "\x1b[31mERROR\x1b[0m" // 红色
	case zapcore.FatalLevel:
		level = "\x1b[35mFATAL\x1b[0m" // 紫色
	default:
		level = fmt.Sprintf("%v", l)
	}
	enc.AppendString(level)
}

// WithFields 添加字段到日志
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	return sugar.With(fieldsToArgs(fields)...)
}

func fieldsToArgs(fields map[string]interface{}) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return args
}

// Debug 输出 Debug 级别日志
func Debug(args ...interface{}) {
	sugar.Debug(args...)
}

// Info 输出 Info 级别日志
func Info(args ...interface{}) {
	sugar.Info(args...)
}

// Warn 输出 Warn 级别日志
func Warn(args ...interface{}) {
	sugar.Warn(args...)
}

// Error 输出 Error 级别日志
func Error(args ...interface{}) {
	sugar.Error(args...)
}

// Fatal 输出 Fatal 级别日志
func Fatal(args ...interface{}) {
	sugar.Fatal(args...)
}

// Debugf 输出 Debug 级别格式化日志
func Debugf(format string, args ...interface{}) {
	sugar.Debugf(format, args...)
}

// Infof 输出 Info 级别格式化日志
func Infof(format string, args ...interface{}) {
	sugar.Infof(format, args...)
}

// Warnf 输出 Warn 级别格式化日志
func Warnf(format string, args ...interface{}) {
	sugar.Warnf(format, args...)
}

// Errorf 输出 Error 级别格式化日志
func Errorf(format string, args ...interface{}) {
	sugar.Errorf(format, args...)
}

// Fatalf 输出 Fatal 级别格式化日志
func Fatalf(format string, args ...interface{}) {
	sugar.Fatalf(format, args...)
}

// Sync 同步日志缓冲
func Sync() error {
	return log.Sync()
}
