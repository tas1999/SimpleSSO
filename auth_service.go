package main

import (
	"SimpleSSO/repository"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
)

type AuthService struct {
	db     *repository.Repository
	secret string
}
type LoginDto struct {
	RefreshToken repository.RefreshToken
	Token        Token
}
type Token struct {
	Token      string
	Expiration int64
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
	hash, err := a.GetHash(user.Password)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	if dbUser.Password != hash {
		fmt.Fprint(w, "invalid password")
		return
	}
	loginDto, err := a.GetLoginData(dbUser.Id)
	w.Header().Add("Content-Type", "application/json")
	jen := json.NewEncoder(w)
	err = jen.Encode(loginDto)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
}
func (a *AuthService) RefreshToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	user := repository.RefreshToken{}
	err := decoder.Decode(&user)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	rt, err := a.db.GetRefreshToken(user.Token)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	if rt.Expiration < time.Now().UnixMilli() {
		fmt.Fprint(w, "expiration has expired")
		return
	}
	if rt.UserId != user.UserId {
		fmt.Fprint(w, "token isn't valid")
		return
	}
	loginDto, err := a.GetLoginData(user.UserId)
	w.Header().Add("Content-Type", "application/json")
	jen := json.NewEncoder(w)
	err = jen.Encode(loginDto)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
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
		fmt.Fprint(w, "invalid data")
		return
	}
	hash, err := a.GetHash(user.Password)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	userdb := repository.User{
		Login:    user.Login,
		Password: hash,
	}
	userRes, err := a.db.SetUser(userdb)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	w.Header().Add("Content-Type", "application/json")
	loginDto, err := a.GetLoginData(userRes.Id)
	jen := json.NewEncoder(w)
	err = jen.Encode(loginDto)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
}
func (a *AuthService) GetLoginData(userId int) (*LoginDto, error) {
	exp := time.Now().Add(time.Hour * 24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(exp),
			IssuedAt:  jwt.At(time.Now()),
		},
		Id: uuid.New().String(),
	})
	tokenSign, err := token.SignedString([]byte("test"))
	if err != nil {
		return nil, err
	}
	loginDto := LoginDto{}
	loginDto.Token = Token{
		Token:      tokenSign,
		Expiration: exp.UnixMilli(),
	}
	rt := repository.RefreshToken{
		UserId:     userId,
		Token:      "test",
		Expiration: time.Now().Add(time.Hour * 24 * 30).UnixMilli(),
	}
	rtRes, err := a.db.SetRefreshToken(rt)
	if err != nil {
		return nil, err
	}
	loginDto.RefreshToken = *rtRes
	return &loginDto, nil
}

func (a *AuthService) GetHash(password string) (string, error) {
	hash := crypto.SHA256.New()
	_, err := hash.Write([]byte(password))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(hash.Sum([]byte(a.secret))), nil
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
