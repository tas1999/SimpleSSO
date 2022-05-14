package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
)

func main() {
	db, err := New()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	auth := AuthService{db: db}
	http.HandleFunc("/login", auth.Login)
	http.HandleFunc("/registration", auth.Registration)
	s := &http.Server{
		Addr: ":8080",
	}
	err = s.ListenAndServe()
	fmt.Println(err.Error())
}

type AuthService struct {
	db *Repository
}

func (a *AuthService) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	user := UserLogin{}
	err := decoder.Decode(&user)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	dbUser, err := a.db.GetUser(user.Login)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	if dbUser.Password != user.Password {
		fmt.Fprint(w, "invalid password")
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.At(time.Now()),
		},
		Username: user.Login,
		Id:       uuid.New().String(),
	})
	tokenSign, err := token.SignedString([]byte("test"))
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, tokenSign)
}
func (a *AuthService) Registration(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	user := UserLogin{}
	err := decoder.Decode(&user)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	_, err = a.db.GetUser(user.Login)
	if err == nil {
		fmt.Fprint(w, "InvalidData")
		return
	}
	userdb := User{
		Login:    user.Login,
		Password: user.Password,
	}
	userRes, err := a.db.SetUser(userdb)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Println(userRes)
	enc := json.NewEncoder(w)
	err = enc.Encode(userRes)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
}

type UserLogin struct {
	Login    string
	Password string
}
type Claims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Id       string `json:"sub"`
}
