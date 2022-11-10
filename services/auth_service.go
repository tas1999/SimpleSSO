package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tas1999/SimpleSSO/repository"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
}
type RefreshService interface {
	GetRefreshToken(token string) (*repository.RefreshToken, error)
	SetRefreshToken(token repository.RefreshToken) (*repository.RefreshToken, error)
}
type AuthService struct {
	db     Repository
	rdb    RefreshService
	crypt  Crypt
	logger *logr.Logger
}
type LoginDto struct {
	RefreshToken repository.RefreshToken
	Token        Token
}
type Token struct {
	Token      string
	Expiration int64
}

func New(db Repository, rdb RefreshService, crypt Crypt, logger *logr.Logger) (*AuthService, error) {
	return &AuthService{
		db:     db,
		rdb:    rdb,
		crypt:  crypt,
		logger: logger,
	}, nil
}
func (a *AuthService) LoginHttp(w http.ResponseWriter, r *http.Request) {
	opsLogin.Inc()
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	user := UserLogin{}
	err := decoder.Decode(&user)
	if err != nil {
		a.WriteError(w, err)
		return
	}
	loginDto, err := a.login(user)
	if err != nil {
		a.WriteError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	jen := json.NewEncoder(w)
	err = jen.Encode(loginDto)
	if err != nil {
		a.WriteError(w, err)
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
	opsRefreshToken.Inc()
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	user := repository.RefreshToken{}
	err := decoder.Decode(&user)
	if err != nil {
		a.WriteError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	jen := json.NewEncoder(w)
	loginDto, err := a.refreshToken(user)
	if err != nil {
		a.WriteError(w, err)
		return
	}
	err = jen.Encode(loginDto)
	if err != nil {
		a.WriteError(w, err)
		return
	}
}
func (a *AuthService) refreshToken(user repository.RefreshToken) (*LoginDto, error) {
	rt, err := a.rdb.GetRefreshToken(user.Token)
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
		a.WriteError(w, err)
		return
	}
	loginDto, err := a.registration(user)
	if err != nil {
		a.WriteError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	jen := json.NewEncoder(w)
	err = jen.Encode(loginDto)
	if err != nil {
		a.WriteError(w, err)
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
	rtRes, err := a.rdb.SetRefreshToken(rt)
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

func (a *AuthService) WriteError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	fmt.Fprint(w, err.Error())
}

var (
	opsLogin = promauto.NewCounter(prometheus.CounterOpts{
		Name: "simplesso_login_ops_total",
		Help: "The total number of loging events",
	})
	opsRefreshToken = promauto.NewCounter(prometheus.CounterOpts{
		Name: "simplesso_refresh_token_ops_total",
		Help: "The total number of refresh token events",
	})
)
