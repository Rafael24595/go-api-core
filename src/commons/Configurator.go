package commons

import (
	"os"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/dependency"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func Initialize(kargs map[string]utils.Any) (*configuration.Configuration, *dependency.DependencyContainer) {
	session := uuid.NewString()
	timestamp := time.Now().UnixMilli()

	log.ConfigureLog(session, timestamp, kargs)

	mod := ReadGoMod()
	pkg := ReadPackage()
	config := configuration.Initialize(session, timestamp, kargs, mod, &pkg.Project)

	log.Messagef("Session ID: %s", config.SessionId())
	log.Messagef("Started at: %s", utils.FormatMilliseconds(config.Timestamp()))
	log.Messagef("Dev mode: %v", config.Dev())

	container := dependency.Initialize()
	initializeManagerSession(container)
	return &config, container
}

func initializeManagerSession(container *dependency.DependencyContainer) *repository.ManagerSession {
	file := repository.NewManagerCsvtFile(dto.NewDtoSessionDefault, repository.CSVT_FILE_PATH_SESSION)
	manager, err := repository.InitializeManagerSession(file, container.ManagerCollection, container.ManagerGroup)
	if err != nil {
		log.Panic(err)
	}

	return manager
}

func ReadEnv(file string) map[string]utils.Any {
	if len(file) > 0 {
		if err := godotenv.Load(".env"); err != nil {
			log.Warningf("Error during environment loading file from '%s'", file)
		}
	}

	envs := make(map[string]utils.Any)
	for _, env := range os.Environ() {
		pair := splitEnv(env)
		envs[pair[0]] = *utils.AnyFrom(pair[1])
	}

	return envs
}

func splitEnv(env string) []string {
	var pair []string
	for i, char := range env {
		if char == '=' {
			pair = append(pair, env[:i], env[i+1:])
			break
		}
	}
	return pair
}

func ReadGoMod() *configuration.Mod {
	file, err := os.Open("go.mod")
	if err != nil {
		log.Panic(err)
	}

	result := configuration.DecodeMod(file)
	err = file.Close()
	if err != nil {
		log.Panic(err)
	}

	return result
}

func ReadPackage() *configuration.Package {
	file, err := os.Open("go.package.yml")
	if err != nil {
		log.Panicf("Error opening go.package.yml: %v", err)
	}

	var pkg configuration.Package
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&pkg); err != nil {
		log.Panicf("Error decoding YAML: %v", err)
	}

	if err := file.Close(); err != nil {
		log.Panicf("Error closing file: %v", err)
	}

	return &pkg
}
