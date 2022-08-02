package datainvok

import (
	"SimpleSSO/provider"
	"fmt"
	"os"
	"plugin"
)

//type ProviderFactory func(conf map[string]interface{}) (Provider, error)
type ProviderFactory interface {
	GetProvider(conf map[string]interface{}) (provider.Provider, error)
}
type DateInvoker struct {
	providers map[string]provider.Provider
}
type DateInvokerConfig struct {
	ProvidersConfig map[string]ProviderConfig
	Enable          bool
}
type ProviderConfig struct {
	Path   string
	Config map[string]interface{}
}

func New(dic DateInvokerConfig) (*DateInvoker, error) {
	if !dic.Enable {
		return &DateInvoker{}, nil
	}
	dateInv := DateInvoker{
		providers: make(map[string]provider.Provider, len(dic.ProvidersConfig)),
	}
	for n, v := range dic.ProvidersConfig {
		pr, err := loadProvider(v)
		if err != nil {
			return nil, err
		}
		dateInv.providers[n] = pr
	}
	return &dateInv, nil

}
func loadProvider(pc ProviderConfig) (provider.Provider, error) {
	if _, err := os.Stat(pc.Path); os.IsNotExist(err) {
		return nil, err
	}
	plug, err := plugin.Open(pc.Path)
	if err != nil {
		return nil, err
	}
	n, err := plug.Lookup("Factory")
	if err != nil {
		return nil, err
	}
	fmt.Printf("t1: %T\n", n)
	var fac ProviderFactory
	fac, ok := n.(ProviderFactory)
	if !ok {
		return nil, fmt.Errorf("new is not ProviderFactory")
	}
	return fac.GetProvider(pc.Config)
}
func (d *DateInvoker) GetData(userId int) map[string]map[string]interface{} {
	if d.providers == nil {
		return nil
	}
	data := make(map[string]map[string]interface{}, len(d.providers))
	for n, v := range d.providers {
		d, err := v.GetData(userId)
		if err != nil {
			return nil
		}
		data[n] = d
	}
	return data
}
