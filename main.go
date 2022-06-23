package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"SimpleSSO/cryptos"
	"SimpleSSO/repository"
	"SimpleSSO/services"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Postgres    repository.PostgresConfig `mapstructure:"postgres"`
	Cryptos     cryptos.Secret            `mapstructure:"cryptos"`
	ZapConfig   *ZapConfig                `mapstructure:"zap_config"`
	RedisConfig repository.RedisConfig    `mapstructure:"redis"`
}
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
	defer zapLog.Sync()
	if err != nil {
		return nil, err
	}
	logger = zapr.NewLogger(zapLog)
	return &logger, nil
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "__")
	viper.SetEnvKeyReplacer(replacer)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("viper read in config error", err.Error())
		return
	}
	conf := &Config{}
	err = viper.Unmarshal(conf)
	if err != nil {
		fmt.Println(context.Background(), "viper unmarshal error", err.Error())
		return
	}
	logger, err := NewLog(conf.ZapConfig)
	if err != nil {
		fmt.Println(context.Background(), "zap log error", err.Error())
		return
	}
	logger.Info("conf", "conf", conf)

	mconf := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		conf.Postgres.Username, conf.Postgres.Password, conf.Postgres.Host, conf.Postgres.Port, conf.Postgres.DBName, conf.Postgres.SSLMode)
	logger.Info(mconf)
	m, err := migrate.New(
		"file://migrations",
		mconf)
	if err != nil {
		logger.Error(err, "create migrations error")
		return
	}
	err = m.Up()
	if err != nil {
		logger.Error(err, "migrations up error")
	}
	db, err := repository.New(conf.Postgres, logger)
	if err != nil {
		logger.Error(err, "create repository error")
		return
	}
	rdb, err := repository.NewRedis(conf.RedisConfig, logger)
	auth, err := services.New(db, rdb, &conf.Cryptos, logger)
	if err != nil {
		logger.Error(err, "create services error")
		return
	}
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/login", auth.LoginHttp)
	http.HandleFunc("/registration", auth.RegistrationHttp)
	http.HandleFunc("/refreshToken", auth.RefreshTokenHttp)
	s := &http.Server{
		Addr: ":8080",
	}
	err = s.ListenAndServe()
	logger.Error(err, "listen and serve error")
}
