package commons

import (
	"maps"
	"os"
	"strings"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/dependency"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func Initialize(kargs map[string]utils.Argument) (*configuration.Configuration, *dependency.DependencyContainer) {
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
	return repository.InitializeManagerSession(file, container.ManagerCollection, container.ManagerGroup)
}

func ReadAllEnv(path string) map[string]utils.Argument {
	envs := ReadDotEnv(path)
	maps.Copy(envs, ReadEnv())
	return envs
}

func ReadDotEnv(path string) map[string]utils.Argument {
	envs := make(map[string]utils.Argument)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return envs
	}

	result, err := os.ReadFile(path)
	if err != nil {
		return envs
	}

	for line := range strings.SplitSeq(string(result), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if key, value, ok := manageEnv(line); ok {
			envs[key] = *value
		}
	}

	return envs
}

func ReadEnv() map[string]utils.Argument {
	envs := make(map[string]utils.Argument)
	for _, env := range os.Environ() {
		if key, value, ok := manageEnv(env); ok {
			envs[key] = *value
		}
	}
	return envs
}

func manageEnv(env string) (string, *utils.Argument, bool) {
	parts := strings.SplitN(env, "=", 2)
	if len(parts) == 2 {
		return parts[0], utils.ArgumentFrom(parts[1]), true
	}
	return "", nil, false
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
