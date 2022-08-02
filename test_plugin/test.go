package main

import "SimpleSSO/provider"

var Factory TestProviderFactory

type TestProviderFactory struct {
}

func (f *TestProviderFactory) GetProvider(conf map[string]interface{}) (provider.Provider, error) {
	return &TestProvider{}, nil
}

type TestProvider struct {
}

func (p *TestProvider) GetData(userId int) (map[string]interface{}, error) {
	return map[string]interface{}{
		"brand": "ru",
	}, nil
}
