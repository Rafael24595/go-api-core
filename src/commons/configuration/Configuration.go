package configuration

import (
	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

var instance *Configuration

type Configuration struct {
	admin  string
	secret []byte
	kargs  map[string]utils.Any
}

func Initialize(kargs map[string]utils.Any) Configuration {
	if instance != nil {
		panic("")
	}

	admin, ok := kargs["GO_API_ADMIN_USER"].String()
	if !ok {
		panic("Admin is not defined")
	}

	secret, ok := kargs["GO_API_ADMIN_SECRET"].String()
	if !ok {
		panic("Secret is not defined")
	}

	instance = &Configuration{
		admin:  admin,
		secret: []byte(secret),
		kargs:  kargs,
	}

	return *instance
}

func Instance() Configuration {
	if instance == nil {
		panic("")
	}
	return *instance
}

func (c Configuration) Admin() string {
	return c.admin
}

func (c Configuration) Secret() []byte {
	return c.secret
}
