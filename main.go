package main

import (
	"fmt"
	"net/http"
	"strings"

	"SimpleSSO/cryptos"
	"SimpleSSO/repository"
	"SimpleSSO/services"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
)

type Config struct {
	Postgres repository.PostgresConfig `mapstructure:"postgres"`
	Cryptos  cryptos.Secret            `mapstructure:"cryptos"`
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "__")
	viper.SetEnvKeyReplacer(replacer)
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
	fmt.Println(conf)

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
	db, err := repository.New(conf.Postgres)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	auth, err := services.New(db, &conf.Cryptos)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	http.HandleFunc("/login", auth.LoginHttp)
	http.HandleFunc("/registration", auth.RegistrationHttp)
	http.HandleFunc("/refreshToken", auth.RefreshTokenHttp)
	s := &http.Server{
		Addr: ":8080",
	}
	err = s.ListenAndServe()
	fmt.Println(err.Error())
}
