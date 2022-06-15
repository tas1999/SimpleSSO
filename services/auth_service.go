package services

import (
	"SimpleSSO/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

type Crypt interface {
	GetHash(string) (string, error)
	GetJwt(jwt.Claims) (string, error)
	GenerateSecureToken() string
}
type Repository interface {
	GetUser(login string) (*repository.User, error)
	GetUserById(Id int) (*repository.User, error)
	SetUser(user repository.User) (*repository.User, error)
	GetRefreshToken(token string) (*repository.RefreshToken, error)
	SetRefreshToken(token repository.RefreshToken) (*repository.RefreshToken, error)
}
type AuthService struct {
	db    Repository
	crypt Crypt
}
type LoginDto struct {
	RefreshToken repository.RefreshToken
	Token        Token
}
type Token struct {
	Token      string
	Expiration int64
}

func New(db Repository, crypt Crypt) (*AuthService, error) {
	return &AuthService{
		db:    db,
		crypt: crypt,
	}, nil
}
func (a *AuthService) LoginHttp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	user := UserLogin{}
	err := decoder.Decode(&user)
	if err != nil {
		WriteError(w, err)
		return
	}
	loginDto, err := a.login(user)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	jen := json.NewEncoder(w)
	err = jen.Encode(loginDto)
	if err != nil {
		WriteError(w, err)
		return
	}
}
func (a *AuthService) login(user UserLogin) (*LoginDto, error) {
	dbUser, err := a.db.GetUser(user.Login)
	if err != nil {
		return nil, err
	}
	hash, err := a.crypt.GetHash(user.Password)
	if err != nil {
		return nil, err
	}
	if dbUser.Password != hash {
		return nil, fmt.Errorf("invalid password")
	}
	return a.GetLoginData(dbUser.Id)

}
func (a *AuthService) RefreshTokenHttp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	user := repository.RefreshToken{}
	err := decoder.Decode(&user)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	jen := json.NewEncoder(w)
	loginDto, err := a.refreshToken(user)
	if err != nil {
		WriteError(w, err)
		return
	}
	err = jen.Encode(loginDto)
	if err != nil {
		WriteError(w, err)
		return
	}
}
func (a *AuthService) refreshToken(user repository.RefreshToken) (*LoginDto, error) {
	rt, err := a.db.GetRefreshToken(user.Token)
	if err != nil {
		return nil, err
	}
	if rt.Expiration < time.Now().UnixMilli() {
		return nil, fmt.Errorf("expiration has expired")
	}
	if rt.UserId != user.UserId {
		return nil, fmt.Errorf("token isn't valid")
	}
	return a.GetLoginData(user.UserId)

}
func (a *AuthService) RegistrationHttp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	user := UserLogin{}
	err := decoder.Decode(&user)
	if err != nil {
		WriteError(w, err)
		return
	}
	loginDto, err := a.registration(user)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	jen := json.NewEncoder(w)
	err = jen.Encode(loginDto)
	if err != nil {
		WriteError(w, err)
		return
	}
}
func (a *AuthService) registration(user UserLogin) (*LoginDto, error) {
	_, err := a.db.GetUser(user.Login)
	if err == nil {
		return nil, fmt.Errorf("invalid data")
	}
	hash, err := a.crypt.GetHash(user.Password)
	if err != nil {
		return nil, err
	}
	userdb := repository.User{
		Login:    user.Login,
		Password: hash,
	}
	userRes, err := a.db.SetUser(userdb)
	if err != nil {
		return nil, err
	}

	return a.GetLoginData(userRes.Id)
}
func (a *AuthService) GetLoginData(userId int) (*LoginDto, error) {
	exp := time.Now().Add(time.Hour * 24)
	tokenSign, err := a.crypt.GetJwt(&Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(exp),
			IssuedAt:  jwt.At(time.Now()),
		},
		Id: fmt.Sprint(userId),
	})
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
		Token:      a.crypt.GenerateSecureToken(),
		Expiration: time.Now().Add(time.Hour * 24 * 30).UnixMilli(),
	}
	rtRes, err := a.db.SetRefreshToken(rt)
	if err != nil {
		return nil, err
	}
	loginDto.RefreshToken = *rtRes
	return &loginDto, nil
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

func WriteError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	fmt.Fprint(w, err.Error())
}
