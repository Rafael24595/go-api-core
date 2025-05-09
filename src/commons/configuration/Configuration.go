package configuration

import (
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

var instance *Configuration

type Configuration struct {
	Mod     Mod
	Project Project
	admin   string
	secret  []byte
	kargs   map[string]utils.Any
}

func Initialize(kargs map[string]utils.Any, mod *Mod, project *Project) Configuration {
	if instance != nil {
		log.Panics("The configuration is alredy initialized")
	}

	admin, ok := kargs["GO_API_ADMIN_USER"].String()
	if !ok {
		log.Panics("Admin is not defined")
	}

	secret, ok := kargs["GO_API_ADMIN_SECRET"].String()
	if !ok {
		log.Panics("Secret is not defined")
	}

	instance = &Configuration{
		Mod:     *mod,
		Project: *project,
		admin:   admin,
		secret:  []byte(secret),
		kargs:   kargs,
	}

	return *instance
}

func Instance() Configuration {
	if instance == nil {
		log.Panics("The configuration is not initialized yet")
	}
	return *instance
}

func (c Configuration) Admin() string {
	return c.admin
}

func (c Configuration) Secret() []byte {
	return c.secret
}
