package provider

type Provider interface {
	GetData(userId int) (map[string]interface{}, error)
}
