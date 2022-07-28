package main

import (
	"context"
	"fmt"
	"net/http"

	"SimpleSSO/config"
	"SimpleSSO/logger"
	"SimpleSSO/repository"
	"SimpleSSO/services"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	conf, err := config.GetConfig()
	if err != nil {
		fmt.Println("viper read in config error", err.Error())
		return
	}
	logger, err := logger.NewLog(conf.ZapConfig)
	if err != nil {
		fmt.Println(context.Background(), "zap log error", err.Error())
		return
	}
	err = migrateUp(conf.Migrations)
	if err != nil {
		logger.Error(err, "migrate up error")
	}
	db, err := repository.New(conf.Postgres, logger)
	if err != nil {
		logger.Error(err, "create repository error")
		return
	}
	rdb, err := repository.NewRedis(conf.RedisConfig, logger)
	if err != nil {
		logger.Error(err, "create redis services error")
		return
	}
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
func parseConf(conf repository.PostgresConfig) string {
	mconf := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.DBName, conf.SSLMode)
	return mconf
}
func migrateUp(conf repository.PostgresConfig) error {
	m, err := migrate.New(
		"file://migrations",
		parseConf(conf))
	if err != nil {
		return err
	}
	err = m.Up()
	return err
}
