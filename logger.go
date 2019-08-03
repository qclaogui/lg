package lg

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// APPLog is global logger
	APPLog *zap.Logger

	// TimeFormat is custom Time format
	// example: "2006-01-02T15:04:05.999999999Z07:00"
	// 推荐不要设置, 使用默认时间戳
	TimeFormat string

	// onceInit guarantee initialize logger only once
	onceInit sync.Once
)

type commonInfo struct {
	Project  string `json:"project"`
	Hostname string `json:"hostname"`
}

func (i *commonInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("project", i.Project)
	enc.AddString("hostname", i.Hostname)
	return nil
}

// customTimeEncoder encode Time to our custom format
// This example how we can customize zap default functionality
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(TimeFormat))
}

// Init initializes log by input parameters
// lvl - global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
// timeFormat - custom time format for logger of empty string to use default
func Init(lvl int, project string) (err error) {
	onceInit.Do(func() {
		// First, define our level-handling logic.
		globalLevel := zapcore.Level(lvl)
		// High-priority output should also go to standard error, and low-priority
		// output should also go to standard out.
		// It is usefull for Kubernetes deployment.
		// Kubernetes interprets os.Stdout log items as INFO and os.Stderr log items
		// as ERROR by default.
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= globalLevel && lvl < zapcore.ErrorLevel
		})

		cfg := zap.NewProductionEncoderConfig()
		if len(TimeFormat) > 0 {
			// 默认的TimeKey为(ts) float64类型. 自定义会将TimeKey修改为string,防止es中出现问题,所以换个新的key叫tsp
			cfg.TimeKey = "tsp"
			cfg.EncodeTime = customTimeEncoder
		}
		// Optimize the Kafka output for machine consumption and the console output
		// for human operators.
		consoleEncoder := zapcore.NewJSONEncoder(cfg)
		// Join the outputs, encoders, and level-handling functions into
		// zapcore.Cores, then tee the four cores together.
		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), lowPriority),
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stderr), highPriority),
		)

		// ErrorLevel 堆栈跟踪
		stackTrace := zap.AddStacktrace(zap.ErrorLevel)
		// 设置初始化字段
		fields := zap.Fields(zap.Object("info", &commonInfo{project, getHostName()}))

		// From a zapcore.Core, it's easy to construct a Logger.
		APPLog = zap.New(core, fields, stackTrace)
		zap.RedirectStdLog(APPLog)
	})

	return err
}

// getHostName 获取当前主机的Hostname
func getHostName() string {
	if host, err := os.Hostname(); err == nil {
		return host
	}
	return "unknown"
}
