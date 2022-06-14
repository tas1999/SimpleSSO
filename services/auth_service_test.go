package services

import (
	"SimpleSSO/repository"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestLogin_200(t *testing.T) {
	cont := gomock.NewController(t)
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
	auth, err := New(rep, cr)
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
