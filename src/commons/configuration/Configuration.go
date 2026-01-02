package configuration

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/commons/format"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

var (
	instance *Configuration
	once     sync.Once
)

type Snapshot struct {
	Enable bool
	Time   int64
	Limit  int
}

type Configuration struct {
	Signal    *signalHandler
	EventHub  *system.SystemEventHub
	Mod       Mod
	Project   Project
	dev       bool
	sessionId string
	timestamp int64
	admin     string
	secret    []byte
	format    format.DataFormat
	snapshot  Snapshot
	kargs     map[string]utils.Argument
}

func Initialize(session string, timestamp int64, kargs map[string]utils.Argument, mod *Mod, project *Project, snapshot *Snapshot) Configuration {
	once.Do(func() {
		admin := kargs["GAC_ADMIN_USER"].String()
		if admin == "" {
			log.Panics("Admin username is not defined")
		}

		secret := kargs["GAC_ADMIN_SECRET"].String()
		if secret == "" {
			log.Panics("Admin secret is not defined")
		}

		dev := kargs["GAC_DEV"].Boold(false)

		instance = &Configuration{
			Signal:    newSignalHandler(),
			EventHub:  system.InitializeSystemEventHub(),
			Mod:       *mod,
			Project:   *project,
			format:    format.CSVT,
			dev:       dev,
			sessionId: session,
			timestamp: timestamp,
			admin:     admin,
			secret:    []byte(secret),
			snapshot:  *snapshot,
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

func (c Configuration) Format() format.DataFormat {
	return c.format
}

func (c Configuration) Snapshot() Snapshot {
	return c.snapshot
}
