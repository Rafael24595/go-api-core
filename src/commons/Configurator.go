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
	"gopkg.in/yaml.v3"
)

func Initialize(kargs map[string]utils.Any) (*configuration.Configuration, *dependency.DependencyContainer) {
	mod := ReadGoMod()
	pkg := ReadPackage()
	config := configuration.Initialize(kargs, mod, &pkg.Project)
	container := dependency.Initialize()
	initializeManagerSession(container)
	return &config, container
}

func initializeManagerSession(container *dependency.DependencyContainer) *repository.ManagerSession {
	file := repository.NewManagerCsvtFile(dto.NewDtoSessionDefault, repository.CSVT_FILE_PATH_SESSION)
	manager, err := repository.InitializeManagerSession(file, container.ManagerCollection, container.ManagerGroup)
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

func ReadGoMod() *configuration.Mod {
	file, err := os.Open("go.mod")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	return configuration.DecodeMod(file)
}

func ReadPackage() *configuration.Package {
	file, err := os.Open("go.package.yml")
	if err != nil {
		panic(fmt.Sprintf("Error opening go.package.yml: %v", err))
	}
	defer file.Close()

	var pkg configuration.Package
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&pkg); err != nil {
		panic(fmt.Sprintf("Error decoding YAML: %v", err))
	}

	return &pkg
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
