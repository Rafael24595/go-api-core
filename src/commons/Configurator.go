package commons

import (
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/dependency"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

func Initialize(kargs map[string]string) (*configuration.Configuration, *dependency.DependencyContainer) {
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
