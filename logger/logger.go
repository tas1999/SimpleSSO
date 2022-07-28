package logger

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapConfig struct {
	Encoding string `mapstructure:"encoding"`
	Level    int8   `mapstructure:"level"`
}

func NewLog(conf *ZapConfig) (*logr.Logger, error) {
	var logger logr.Logger
	if conf == nil {
		conf = &ZapConfig{
			Level:    int8(zapcore.InfoLevel),
			Encoding: "json",
		}
	}
	atom := zap.NewAtomicLevelAt(zapcore.Level(conf.Level))
	zapConf := zap.NewProductionConfig()
	zapConf.Encoding = conf.Encoding
	zapConf.Level = atom
	zapLog, err := zapConf.Build()
	if err != nil {
		return nil, err
	}
	logger = zapr.NewLogger(zapLog)
	return &logger, nil
}
