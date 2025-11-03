package dependency

import (
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	collection_domain "github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
	token_domain "github.com/Rafael24595/go-api-core/src/domain/token"
	"github.com/Rafael24595/go-api-core/src/infrastructure"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	repository_collection "github.com/Rafael24595/go-api-core/src/infrastructure/repository/collection"
	repository_context "github.com/Rafael24595/go-api-core/src/infrastructure/repository/context"
	repository_group "github.com/Rafael24595/go-api-core/src/infrastructure/repository/group"
	repository_mock "github.com/Rafael24595/go-api-core/src/infrastructure/repository/mock"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/request"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/response"
	repository_token "github.com/Rafael24595/go-api-core/src/infrastructure/repository/token"
	"github.com/Rafael24595/go-collections/collection"
)

const (
	PRESIST_PREFIX = "sve"
	HISTORY_PREFIX = "hst"
)

var instance *DependencyContainer

type DependencyContainer struct {
	RepositoryContext repository.IRepositoryContext
	ManagerRequest    *repository.ManagerRequest
	ManagerContext    *repository.ManagerContext
	ManagerCollection *repository.ManagerCollection
	ManagerHistoric   *repository.ManagerHistoric
	ManagerGroup      *repository.ManagerGroup
	ManagerEndPoint   *repository.ManagerEndPoint
	ManagerToken      *repository.ManagerToken
}

func Initialize() *DependencyContainer {
	if instance != nil {
		log.Panics("The dependency container is alredy initialized")
	}

	_, err := infrastructure.WarmUp()
	if err != nil {
		log.Error(err)
	}

	repositoryRequest := loadRepositoryRequest()
	repositoryResponse := loadRepositoryResponse()

	repositoryContext := loadRepositoryContext()
	repositoryCollection := loadRepositoryCollection()
	repositoryGroup := loadRepositoryGroup()
	repositoryEndPoint := loadRepositoryEndPoint()
	repositoryToken := loadRepositoryToken()

	managerRequest := loadManagerRequest(repositoryRequest, repositoryResponse)
	managerContext := loadManagerContext(repositoryContext)
	managerCollection := loadManagerCollection(repositoryCollection, managerContext, managerRequest)
	managerHistoric := loadManagerHistoric(managerRequest, managerCollection)
	managerGroup := loadManagerGroup(repositoryGroup, managerCollection)
	managerEndPoint := loadManagerEndPoint(repositoryEndPoint)
	managerToken := loadManagerToken(repositoryToken)

	container := &DependencyContainer{
		RepositoryContext: repositoryContext,
		ManagerRequest:    managerRequest,
		ManagerContext:    managerContext,
		ManagerCollection: managerCollection,
		ManagerHistoric:   managerHistoric,
		ManagerGroup:      managerGroup,
		ManagerEndPoint:   managerEndPoint,
		ManagerToken:      managerToken,
	}

	instance = container

	return instance
}

func loadRepositoryRequest() repository.IRepositoryRequest {
	file := repository.NewManagerCsvtFile[action.Request](repository.CSVT_FILE_PATH_REQUEST)
	impl := collection.DictionarySyncEmpty[string, action.Request]()
	repository, err := request.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryResponse() repository.IRepositoryResponse {
	file := repository.NewManagerCsvtFile[action.Response](repository.CSVT_FILE_PATH_RESPONSE)
	impl := collection.DictionarySyncEmpty[string, action.Response]()
	repository, err := response.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryContext() repository.IRepositoryContext {
	file := repository.NewManagerCsvtFile[dto.DtoContext](repository.CSVT_FILE_PATH_CONTEXT)
	impl := collection.DictionarySyncEmpty[string, context.Context]()
	repository, err := repository_context.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryCollection() repository.IRepositoryCollection {
	file := repository.NewManagerCsvtFile[collection_domain.Collection](repository.CSVT_FILE_PATH_COLLECTION)
	impl := collection.DictionarySyncEmpty[string, collection_domain.Collection]()
	repository, err := repository_collection.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryGroup() repository.IRepositoryGroup {
	file := repository.NewManagerCsvtFile[domain.Group](repository.CSVT_FILE_PATH_GROUP)
	impl := collection.DictionarySyncEmpty[string, domain.Group]()
	repository, err := repository_group.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryEndPoint() repository.IRepositoryEndPoint {
	file := repository.NewManagerCsvtFile[mock_domain.EndPoint](repository.CSVT_FILE_PATH_END_POINT)
	impl := collection.DictionarySyncEmpty[string, mock_domain.EndPoint]()
	repository, err := repository_mock.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryToken() repository.IRepositoryToken {
	file := repository.NewManagerCsvtFile[token_domain.Token](repository.CSVT_FILE_PATH_TOKEN)
	impl := collection.DictionarySyncEmpty[string, token_domain.Token]()
	repository, err := repository_token.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadManagerRequest(request repository.IRepositoryRequest, response repository.IRepositoryResponse) *repository.ManagerRequest {
	return repository.NewManagerRequest(request, response)
}

func loadManagerContext(context repository.IRepositoryContext) *repository.ManagerContext {
	return repository.NewManagerContext(context)
}

func loadManagerCollection(collection repository.IRepositoryCollection, managerContext *repository.ManagerContext, managerRequest *repository.ManagerRequest) *repository.ManagerCollection {
	return repository.NewManagerCollection(collection, managerContext, managerRequest)
}

func loadManagerHistoric(managerRequest *repository.ManagerRequest, managerCollection *repository.ManagerCollection) *repository.ManagerHistoric {
	return repository.NewManagerHistoric(managerRequest, managerCollection)
}

func loadManagerGroup(group repository.IRepositoryGroup, managerCollection *repository.ManagerCollection) *repository.ManagerGroup {
	return repository.NewManagerGroup(group, managerCollection)
}

func loadManagerEndPoint(endPoint repository.IRepositoryEndPoint) *repository.ManagerEndPoint {
	return repository.NewManagerEndPoint(endPoint)
}

func loadManagerToken(token repository.IRepositoryToken) *repository.ManagerToken {
	return repository.NewManagerToken(token)
}
