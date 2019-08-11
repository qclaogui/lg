package lg

import (
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Shopify/sarama"
)

// kafkaConfig  contains information
type kafkaConfig struct {
	Hosts string
	Topic string
}

// write2Kafka is used for sending logs to kafka.
type write2Kafka struct {
	config       *kafkaConfig
	syncProducer sarama.SyncProducer
}

func (w *write2Kafka) Write(b []byte) (n int, err error) {
	if _, _, err = w.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: w.config.Topic,
		Value: sarama.ByteEncoder(b),
	}); err != nil {
		return
	}
	n = len(b)
	return
}

func newLog2Kafka(cfg *kafkaConfig) (*write2Kafka, error) {
	config := sarama.NewConfig()
	// SyncProducer
	config.Producer.RequiredAcks = sarama.WaitForLocal     // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy // Compress messages
	// Producer.Return.Successes must be true to be used in a SyncProducer
	config.Producer.Return.Successes = true

	var brokerList []string
	for _, broker := range strings.Split(cfg.Hosts, ",") {
		if strings.Index(broker, ":") == -1 {
			broker += ":9092"
		}
		brokerList = append(brokerList, broker)
	}

	var producer sarama.SyncProducer
	var err error
	if producer, err = sarama.NewSyncProducer(brokerList, config); err != nil {
		return nil, err
	}

	return &write2Kafka{config: cfg, syncProducer: producer}, nil
}

// Init initializes log by input parameters
// lvl - global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
// timeFormat - custom time format for logger of empty string to use default
func InitOnlyKafka(lvl int, project, kafkaTopic, brokers string) (err error) {
	onceInit.Do(func() {
		globalLevel := zapcore.Level(lvl)

		KafkaPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= globalLevel
		})
		var ws io.Writer
		if ws, err = newLog2Kafka(&kafkaConfig{Hosts: brokers, Topic: kafkaTopic}); err != nil {
			return
		}

		// Configure console output.
		cfg := zap.NewProductionEncoderConfig()
		if len(TimeFormat) > 0 {
			cfg.TimeKey = "tsp"
			cfg.EncodeTime = customTimeEncoder
		}

		// Optimize the Kafka output for machine consumption and the console output
		// for human operators.
		core := zapcore.NewTee(zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.Lock(zapcore.AddSync(ws)), KafkaPriority))

		// ErrorLevel 堆栈跟踪
		stackTrace := zap.AddStacktrace(zap.ErrorLevel)
		// 设置初始化字段
		fields := zap.Fields(zap.Object("info", &commonInfo{project, getHostName()}))

		// From a zapcore.Core, it's easy to construct a Logger.
		APPLog = zap.New(core, fields, stackTrace)
	})
	return err
}

// Init initializes log by input parameters
// lvl - global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
// timeFormat - custom time format for logger of empty string to use default
func InitWithKafka(lvl int, project, kafkaTopic, brokers string) (err error) {
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

		KafkaPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= globalLevel
		})
		// Assume that we have clients for two Kafka topics. The clients implement
		// zapcore.WriteSyncer and are safe for concurrent use. (If they only
		// implement io.Writer, we can use zapcore.AddSync to add a no-op Sync
		// method. If they're not safe for concurrent use, we can add a protecting
		// mutex with zapcore.Lock.)
		var ws io.Writer
		if ws, err = newLog2Kafka(&kafkaConfig{Hosts: brokers, Topic: kafkaTopic}); err != nil {
			return
		}

		// Configure console output.
		cfg := zap.NewProductionEncoderConfig()
		if len(TimeFormat) > 0 {
			cfg.TimeKey = "tsp"
			cfg.EncodeTime = customTimeEncoder
		}

		// Optimize the Kafka output for machine consumption and the console output
		// for human operators.
		kafkaEncoder := zapcore.NewJSONEncoder(cfg)
		consoleEncoder := zapcore.NewJSONEncoder(cfg)
		// Join the outputs, encoders, and level-handling functions into
		// zapcore.Cores, then tee the four cores together.
		core := zapcore.NewTee(
			zapcore.NewCore(kafkaEncoder, zapcore.Lock(zapcore.AddSync(ws)), KafkaPriority),
			// 同时写一份到标准输出
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), lowPriority),
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stderr), highPriority),
		)

		// ErrorLevel 堆栈跟踪
		stackTrace := zap.AddStacktrace(zap.ErrorLevel)
		// 设置初始化字段
		fields := zap.Fields(zap.Object("info", &commonInfo{project, getHostName()}))

		// From a zapcore.Core, it's easy to construct a Logger.
		APPLog = zap.New(core, fields, stackTrace)
	})
	return err
}
