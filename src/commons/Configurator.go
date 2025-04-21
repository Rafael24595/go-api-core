package commons

import (
	"fmt"
	"os"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/dependency"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/joho/godotenv"
)

func Initialize(kargs map[string]utils.Any) (*configuration.Configuration, *dependency.DependencyContainer) {
	config := configuration.Initialize(kargs)
	container := dependency.Initialize()
	initializeManagerSession(container)
	return &config, container
}

func initializeManagerSession(container *dependency.DependencyContainer) *repository.ManagerSession {
	file := repository.NewManagerCsvtFile(dto.NewDtoSessionDefault, repository.CSVT_FILE_PATH_SESSION)
	manager, err := repository.InitializeManagerSession(file, container.ManagerContext)
	if err != nil {
		panic(err)
	}

	return manager
}

func ReadEnv(file string) map[string]utils.Any {
	if len(file) > 0 {
		if err := godotenv.Load(".env"); err != nil {
			//TODO: Log
			fmt.Printf("Error loading %s file", file)
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