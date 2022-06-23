package services

import (
	"SimpleSSO/cryptos"
	"SimpleSSO/repository"
	"testing"

	"github.com/go-logr/logr/funcr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestLogin_200(t *testing.T) {
	logger := funcr.NewJSON(func(obj string) { t.Log(obj) }, funcr.Options{})
	cont := gomock.NewController(t)
	defer cont.Finish()
	rep := NewMockRepository(cont)
	cr := NewMockCrypt(cont)
	hash := "hash"
	jwt := "jwt"
	rtoken := "rtoken"
	password := "password"
	user := repository.User{
		Login:    "login",
		Password: hash,
		Id:       1,
	}
	rftoken := repository.RefreshToken{
		Id:     1,
		Token:  rtoken,
		UserId: user.Id,
	}
	cr.EXPECT().GetHash(password).Return(hash, nil)
	cr.EXPECT().GetJwt(gomock.Any()).Return(jwt, nil)
	cr.EXPECT().GenerateSecureToken().Return(rtoken)
	rep.EXPECT().GetUser(user.Login).Return(&user, nil)
	rep.EXPECT().SetRefreshToken(gomock.Any()).Return(&rftoken, nil)
	auth, err := New(rep, rep, cr, &logger)
	assert.Nil(t, err)

	lg, err := auth.login(UserLogin{
		Login:    user.Login,
		Password: password,
	})
	assert.Nil(t, err)
	assert.Equal(t, jwt, lg.Token.Token)
	assert.Equal(t, rtoken, lg.RefreshToken.Token)
	assert.Equal(t, rftoken.UserId, lg.RefreshToken.UserId)
	assert.Equal(t, rftoken.Id, lg.RefreshToken.Id)
}
func BenchmarkLogin(b *testing.B) {
	logger := funcr.NewJSON(func(obj string) { b.Log(obj) }, funcr.Options{})
	cont := gomock.NewController(b)
	defer cont.Finish()
	rep := NewMockRepository(cont)
	cr := cryptos.Secret{
		PasswordSecret: "secret",
		JwtSecret:      "secretJwt",
	}

	password := "password"
	hash, err := cr.GetHash(password)
	assert.Nil(b, err)
	rtoken := "rtoken"
	user := repository.User{
		Login:    "login",
		Password: hash,
		Id:       1,
	}
	rftoken := repository.RefreshToken{
		Id:     1,
		Token:  rtoken,
		UserId: user.Id,
	}
	rep.EXPECT().GetUser(user.Login).Return(&user, nil).AnyTimes()
	rep.EXPECT().SetRefreshToken(gomock.Any()).Return(&rftoken, nil).AnyTimes()
	auth, err := New(rep, rep, &cr, &logger)
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = auth.login(UserLogin{
			Login:    user.Login,
			Password: password,
		})
	}
}
