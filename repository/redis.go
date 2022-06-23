package repository

import (
	"context"
	"encoding/json"
	"time"

	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}
type Redis struct {
	rdb    *redis.Client
	logger *logr.Logger
}

func NewRedis(conf RedisConfig, logger *logr.Logger) (*Redis, error) {
	rconf := conf
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", rconf.Host, rconf.Port),
		Password: rconf.Password, // no password set
		DB:       rconf.DB,       // use default DB
	})
	return &Redis{
		rdb:    rdb,
		logger: logger,
	}, nil
}

func (r *Redis) GetRefreshToken(token string) (*RefreshToken, error) {
	ctx := context.Background()
	rc := r.rdb.Get(ctx, "refresh_token_"+token)
	tk, err := rc.Result()
	if err != nil {
		return nil, err
	}
	var rt RefreshToken
	err = json.Unmarshal([]byte(tk), &rt)
	return &rt, err
}
func (r *Redis) SetRefreshToken(token RefreshToken) (*RefreshToken, error) {
	ctx := context.Background()
	json, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}
	rc := r.rdb.Set(ctx, "refresh_token_"+token.Token, string(json), time.Minute)
	err = rc.Err()
	return &token, err
}
