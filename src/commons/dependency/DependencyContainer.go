package dependency

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/infrastructure"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	repository_collection "github.com/Rafael24595/go-api-core/src/infrastructure/repository/collection"
	repository_context "github.com/Rafael24595/go-api-core/src/infrastructure/repository/context"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/historic"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/request"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/response"
	"github.com/Rafael24595/go-collections/collection"
)

const (
	PRESIST_PREFIX = "sve"
	HISTORY_PREFIX = "hst"
)

var instance *DependencyContainer

type DependencyContainer struct {
	RepositoryContext  repository.IRepositoryContext
	RepositoryHistoric repository.IRepositoryHistoric
	ManagerActions     *repository.ManagerRequest
	ManagerContext     *repository.ManagerContext
	ManagerCollection  *repository.ManagerCollection
}

func Initialize() *DependencyContainer {
	if instance != nil {
		panic("//TODO: Already instanced")
	}

	_, err := infrastructure.WarmUp()
	if err != nil {
		println(err.Error())
	}

	repositoryRequest := loadRepositoryRequest()
	repositoryResponse := loadRepositoryResponse()

	repositoryHistoric := loadRepositoryHisotric()
	repositoryContext := loadRepositoryContext()
	repositoryCollection := loadRepositoryCollection()

	managerRequest := loadManagerRequest(repositoryRequest, repositoryResponse)
	managerContext := loadManagerContext(repositoryContext)
	managerCollection := loadManagerCollection(repositoryCollection, repositoryContext, repositoryRequest, repositoryResponse)

	container := &DependencyContainer{
		RepositoryContext:  repositoryContext,
		RepositoryHistoric: repositoryHistoric,
		ManagerActions:     managerRequest,
		ManagerContext:     managerContext,
		ManagerCollection:  managerCollection,
	}

	instance = container

	return instance
}

func loadRepositoryRequest() repository.IRepositoryRequest {
	file := repository.NewManagerCsvtFile(domain.NewRequestDefault, repository.CSVT_FILE_PATH_REQUEST)
	impl := collection.DictionarySyncEmpty[string, domain.Request]()
	repository, err := request.InitializeRepositoryMemory(impl, file)
	if err != nil {
		panic(err)
	}

	return repository
}

func loadRepositoryResponse() repository.IRepositoryResponse {
	file := repository.NewManagerCsvtFile(domain.NewResponseDefault, repository.CSVT_FILE_PATH_RESPONSE)
	impl := collection.DictionarySyncEmpty[string, domain.Response]()
	repository, err := response.InitializeRepositoryMemory(impl, file)
	if err != nil {
		panic(err)
	}

	return repository
}

func loadRepositoryHisotric() repository.IRepositoryHistoric {
	file := repository.NewManagerCsvtFile(domain.NewHistoricDefault, repository.CSVT_FILE_PATH_HISTORIC)
	impl := collection.DictionarySyncEmpty[string, domain.Historic]()
	repository, err := historic.InitializeRepositoryMemory(impl, file)
	if err != nil {
		panic(err)
	}

	return repository
}

func loadRepositoryContext() repository.IRepositoryContext {
	file := repository.NewManagerCsvtFile(dto.NewDtoContextDefault, repository.CSVT_FILE_PATH_CONTEXT)
	impl := collection.DictionarySyncEmpty[string, context.Context]()
	repository, err := repository_context.InitializeRepositoryMemory(impl, file)
	if err != nil {
		panic(err)
	}

	return repository
}

func loadRepositoryCollection() repository.IRepositoryCollection {
	file := repository.NewManagerCsvtFile(domain.NewCollectionDefault, repository.CSVT_FILE_PATH_COLLECTION)
	impl := collection.DictionarySyncEmpty[string, domain.Collection]()
	repository, err := repository_collection.InitializeRepositoryMemory(impl, file)
	if err != nil {
		panic(err)
	}

	return repository
}

func loadManagerRequest(request repository.IRepositoryRequest, response repository.IRepositoryResponse) *repository.ManagerRequest {
	return repository.NewManagerRequest(request, response).
		SetInsertPolicy(fixHistoricSize)
}

func loadManagerContext(context repository.IRepositoryContext) *repository.ManagerContext {
	return repository.NewManagerContext(context)
}

func loadManagerCollection(collection repository.IRepositoryCollection, context repository.IRepositoryContext, request repository.IRepositoryRequest, response repository.IRepositoryResponse) *repository.ManagerCollection {
	return repository.NewManagerCollection(collection, context, request, response)
}

func fixHistoricSize(request *domain.Request, repositoryRequest repository.IRepositoryRequest, repositoryResponse repository.IRepositoryResponse) error {
	//TODO: Implement
	return nil
}
