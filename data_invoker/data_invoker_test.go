package datainvok

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin_200(t *testing.T) {
	prov := "test"
	dic := DateInvokerConfig{
		Enable: true,
		ProvidersConfig: map[string]ProviderConfig{
			prov: {
				Path:   "../plugins/test.so",
				Config: nil,
			},
		},
	}
	di, err := New(dic)
	assert.Nil(t, err)
	data := di.GetData(0)
	d := data[prov]["brand"].(string)
	assert.Equal(t, "ru", d)
}
