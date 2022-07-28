package config

import (
	"SimpleSSO/cryptos"
	"SimpleSSO/logger"
	"SimpleSSO/repository"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Postgres    repository.PostgresConfig `mapstructure:"postgres"`
	Migrations  repository.PostgresConfig `mapstructure:"migrations"`
	Cryptos     cryptos.Secret            `mapstructure:"cryptos"`
	ZapConfig   *logger.ZapConfig         `mapstructure:"zap_config"`
	RedisConfig repository.RedisConfig    `mapstructure:"redis"`
}

func GetConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "__")
	viper.SetEnvKeyReplacer(replacer)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = viper.Unmarshal(conf)
	return conf, err
}
