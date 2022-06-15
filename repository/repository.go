package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"fmt"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"golang.org/x/sync/semaphore"
)

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslMode"`
}

type Repository struct {
	db  *sql.DB
	sem *semaphore.Weighted
	rdb *redis.Client
}

type User struct {
	Id       int
	Login    string
	Password string
}

func New(conf PostgresConfig) (*Repository, error) {
	connStr := fmt.Sprintf("user=%s host=%s port=%d password=%s dbname=%s sslmode=%s",
		conf.Username, conf.Host, conf.Port, conf.Password, conf.DBName, conf.SSLMode)
	db, err := sql.Open("postgres", connStr)
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &Repository{
		db:  db,
		sem: semaphore.NewWeighted(int64(90)),
		rdb: rdb,
	}, err
}

func (r *Repository) GetUser(login string) (*User, error) {
	err := r.sem.Acquire(context.Background(), 1)
	if err != nil {
		return nil, err
	}
	defer r.sem.Release(1)
	rows, err := r.db.Query("select id, login, password from users where login=$1", login)
	if err != nil {
		return nil, err
	}
	u := User{}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&u.Id, &u.Login, &u.Password)
	return &u, err
}
func (r *Repository) GetUserById(Id int) (*User, error) {
	err := r.sem.Acquire(context.Background(), 1)
	if err != nil {
		return nil, err
	}
	defer r.sem.Release(1)
	rows, err := r.db.Query("select id, login, password from users where id=$1", Id)
	if err != nil {
		return nil, err
	}
	u := User{}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&u.Id, &u.Login, &u.Password)
	return &u, err
}
func (r *Repository) SetUser(user User) (*User, error) {
	err := r.sem.Acquire(context.Background(), 1)
	if err != nil {
		return nil, err
	}
	defer r.sem.Release(1)
	row := r.db.QueryRow("INSERT INTO users (login,password) VALUES ($1,$2) returning id", user.Login, user.Password)
	err = row.Err()
	if err != nil {
		fmt.Println("row.Err() SetUser ", err)
		return nil, err
	}
	err = row.Scan(&user.Id)
	if err != nil {
		fmt.Println("Scan err SetUser")
		return nil, err
	}
	return &user, nil
}

type RefreshToken struct {
	Id         int
	UserId     int
	Token      string
	Expiration int64
}

func (r *Repository) GetRefreshTokenOld(token string) (*RefreshToken, error) {
	err := r.sem.Acquire(context.Background(), 1)
	if err != nil {
		return nil, err
	}
	defer r.sem.Release(1)
	rows, err := r.db.Query("select id, user_id, token, expiration from refresh_tokens where token=$1", token)
	if err != nil {
		return nil, err
	}
	rt := RefreshToken{}
	defer rows.Close()
	rows.Next()
	var timeEx time.Time
	err = rows.Scan(&rt.Id, &rt.UserId, &rt.Token, &timeEx)
	rt.Expiration = timeEx.UnixMilli()
	return &rt, err
}
func (r *Repository) GetRefreshToken(token string) (*RefreshToken, error) {
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
func (r *Repository) SetRefreshToken(token RefreshToken) (*RefreshToken, error) {
	ctx := context.Background()
	json, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}
	rc := r.rdb.Set(ctx, "refresh_token_"+token.Token, string(json), time.Minute)
	err = rc.Err()
	return &token, err
}
func (r *Repository) SetRefreshTokenOld(token RefreshToken) (*RefreshToken, error) {
	timeEx := time.UnixMilli(token.Expiration)
	err := r.sem.Acquire(context.Background(), 1)
	if err != nil {
		return nil, err
	}
	defer r.sem.Release(1)
	row := r.db.QueryRow("INSERT INTO refresh_tokens (user_id, token, expiration) VALUES ($1,$2,$3) returning id", token.UserId, token.Token, timeEx)
	err = row.Err()
	if err != nil {
		fmt.Println("row.Err() SetRefreshToken", err)
		return nil, err
	}
	err = row.Scan(&token.Id)
	if err != nil {
		fmt.Println("Scan err")
		return nil, err
	}
	return &token, nil
}
