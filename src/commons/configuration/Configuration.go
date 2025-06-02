package configuration

import (
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

var instance *Configuration

type Configuration struct {
	Signal    *signalHandler
	Mod       Mod
	Project   Project
	dev       bool
	sessionId string
	timestamp int64
	admin     string
	secret    []byte
	kargs     map[string]utils.Any
}

func Initialize(session string, kargs map[string]utils.Any, mod *Mod, project *Project) Configuration {
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

	dev, _ := kargs["GO_API_DEV"].Bool()

	instance = &Configuration{
		Signal:    newSignalHandler(),
		Mod:       *mod,
		Project:   *project,
		dev:       dev,
		sessionId: session,
		timestamp: time.Now().UnixMilli(),
		admin:     admin,
		secret:    []byte(secret),
		kargs:     kargs,
	}

	return *instance
}

func Instance() Configuration {
	if instance == nil {
		log.Panics("The configuration is not initialized yet")
	}
	return *instance
}

func (c Configuration) Dev() bool {
	return c.dev
}

func (c Configuration) SessionId() string {
	return c.sessionId
}

func (c Configuration) Timestamp() int64 {
	return c.timestamp
}

func (c Configuration) Admin() string {
	return c.admin
}

func (c Configuration) Secret() []byte {
	return c.secret
}
