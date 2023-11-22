package zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

type (
	Config struct {
		Name             string   `mapstructure:"name"`
		Format           string   `mapstructure:"format"`
		AddSource        bool     `mapstructure:"addSource"`
		Level            string   `mapstructure:"level"`
		OutputPaths      []string `mapstructure:"outputPaths"`
		ErrorOutputPaths []string `mapstructure:"errorOutputPaths"`
	}
)

func NewHandler(conf Config, opts ...zap.Option) *zapslog.Handler {
	core, err := newCore(conf, opts...)
	if err != nil {
		panic(err)
	}
	return zapslog.NewHandler(core, &zapslog.HandlerOptions{
		LoggerName: conf.Name,
		AddSource:  conf.AddSource,
	})
}

func newCore(conf Config, opts ...zap.Option) (zapcore.Core, error) {
	level, err := zapcore.ParseLevel(conf.Level)
	if err != nil {
		level = zap.DebugLevel
	}
	format := "console"
	if conf.Format == "json" {
		format = conf.Format
	}
	encoder := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	builder := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          format,
		EncoderConfig:     encoder,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	if len(conf.OutputPaths) > 0 && conf.OutputPaths[0] != "" {
		builder.OutputPaths = conf.OutputPaths
	}

	if len(conf.ErrorOutputPaths) > 0 && conf.ErrorOutputPaths[0] != "" {
		builder.ErrorOutputPaths = conf.ErrorOutputPaths
	}
	opts = append(opts, zap.AddCallerSkip(1))
	core, err := builder.Build(opts...)
	if err != nil {
		return nil, err
	}
	return core.Core(), nil
}
