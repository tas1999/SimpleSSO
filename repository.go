package main

import (
	"database/sql"

	"fmt"

	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

type User struct {
	Id       int64
	Login    string
	Password string
}

func New() (*Repository, error) {
	connStr := "user=postgres host=127.0.0.1 port=5430 password=postgres dbname=simplesso sslmode=disable"
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
