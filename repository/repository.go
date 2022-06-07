package repository

import (
	"database/sql"
	"time"

	"fmt"

	_ "github.com/lib/pq"
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
	db *sql.DB
}

type User struct {
	Id       int
	Login    string
	Password string
}

func New(connStr string) (*Repository, error) {
	db, err := sql.Open("postgres", connStr)
	return &Repository{db: db}, err
}

func (r *Repository) GetUser(login string) (*User, error) {
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
	row := r.db.QueryRow("INSERT INTO users (login,password) VALUES ($1,$2) returning id", user.Login, user.Password)
	if row.Err() != nil {
		fmt.Println("row.Err()")
		return nil, row.Err()
	}
	err := row.Scan(&user.Id)
	if err != nil {
		fmt.Println("Scan err")
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

func (r *Repository) GetRefreshToken(token string) (*RefreshToken, error) {
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
func (r *Repository) SetRefreshToken(token RefreshToken) (*RefreshToken, error) {
	timeEx := time.UnixMilli(token.Expiration)
	row := r.db.QueryRow("INSERT INTO refresh_tokens (user_id, token, expiration) VALUES ($1,$2,$3) returning id", token.UserId, token.Token, timeEx)
	if row.Err() != nil {
		fmt.Println("row.Err()")
		return nil, row.Err()
	}
	err := row.Scan(&token.Id)
	if err != nil {
		fmt.Println("Scan err")
		return nil, err
	}
	return &token, nil
}
