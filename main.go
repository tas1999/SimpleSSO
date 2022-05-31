package main

import (
	"fmt"
	"net/http"

	"SimpleSSO/repository"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
)

type Config struct {
	Postgres repository.PostgresConfig `json:"postgres"`
}

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	conf := &Config{}
	err = viper.Unmarshal(conf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	connStr := fmt.Sprintf("user=%s host=%s port=%d password=%s dbname=%s sslmode=%s",
		conf.Postgres.Username, conf.Postgres.Host, conf.Postgres.Port, conf.Postgres.Password, conf.Postgres.DBName, conf.Postgres.SSLMode)
	mconf := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		conf.Postgres.Username, conf.Postgres.Password, conf.Postgres.Host, conf.Postgres.Port, conf.Postgres.DBName, conf.Postgres.SSLMode)
	fmt.Println(mconf)
	m, err := migrate.New(
		"file://migrations",
		mconf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	m.Up()
	db, err := repository.New(connStr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	auth := AuthService{db: db, secret: "secret"}
	http.HandleFunc("/login", auth.Login)
	http.HandleFunc("/registration", auth.Registration)
	http.HandleFunc("/refreshToken", auth.RefreshToken)
	s := &http.Server{
		Addr: ":8080",
	}
	err = s.ListenAndServe()
	fmt.Println(err.Error())
}
