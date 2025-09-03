package configuration

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

var (
	instance *Configuration
	once     sync.Once
)

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

func Initialize(session string, timestamp int64, kargs map[string]utils.Any, mod *Mod, project *Project) Configuration {
	once.Do(func() {
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
			timestamp: timestamp,
			admin:     admin,
			secret:    []byte(secret),
			kargs:     kargs,
		}
	})

	if instance == nil {
		log.Panics("The configuration is not initialized properly")
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
