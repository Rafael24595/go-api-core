package dependency

import (
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/client"
	collection_domain "github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
	token_domain "github.com/Rafael24595/go-api-core/src/domain/token"
	"github.com/Rafael24595/go-api-core/src/infrastructure"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	repository_client "github.com/Rafael24595/go-api-core/src/infrastructure/repository/client"
	repository_collection "github.com/Rafael24595/go-api-core/src/infrastructure/repository/collection"
	repository_context "github.com/Rafael24595/go-api-core/src/infrastructure/repository/context"
	repository_group "github.com/Rafael24595/go-api-core/src/infrastructure/repository/group"
	repository_mock "github.com/Rafael24595/go-api-core/src/infrastructure/repository/mock"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/request"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/response"
	repository_token "github.com/Rafael24595/go-api-core/src/infrastructure/repository/token"
	"github.com/Rafael24595/go-collections/collection"
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
	ManagerMetrics    *repository.ManagerMetrics
	ManagerToken      *repository.ManagerToken
	ManagerClientData *repository.ManagerClientData
}

func Initialize(config configuration.Configuration) *DependencyContainer {
	if instance != nil {
		log.Panics("The dependency container is alredy initialized")
	}

	_, err := infrastructure.WarmUp()
	if err != nil {
		log.Error(err)
	}

	repositoryRequest := loadRepositoryRequest(config)
	repositoryResponse := loadRepositoryResponse(config)

	repositoryContext := loadRepositoryContext(config)
	repositoryCollection := loadRepositoryCollection(config)
	repositoryGroup := loadRepositoryGroup(config)
	repositoryEndPoint := loadRepositoryEndPoint(config)
	repositoryMetrics := loadRepositoryMetrics(config)
	repositoryToken := loadRepositoryToken(config)
	repositoryClient := loadRepositoryClientData(config)

	managerRequest := loadManagerRequest(repositoryRequest, repositoryResponse)
	managerContext := loadManagerContext(repositoryContext)
	managerCollection := loadManagerCollection(repositoryCollection, managerContext, managerRequest)
	managerHistoric := loadManagerHistoric(managerRequest, managerCollection)
	managerGroup := loadManagerGroup(repositoryGroup, managerCollection)
	managerMetrics := loadManagerMetrics(repositoryMetrics)
	managerEndPoint := loadManagerEndPoint(repositoryEndPoint, managerMetrics)
	managerToken := loadManagerToken(repositoryToken)
	managerClientData := loadManagerClientData(repositoryClient, managerCollection, managerGroup)

	container := &DependencyContainer{
		RepositoryContext: repositoryContext,
		ManagerRequest:    managerRequest,
		ManagerContext:    managerContext,
		ManagerCollection: managerCollection,
		ManagerHistoric:   managerHistoric,
		ManagerGroup:      managerGroup,
		ManagerEndPoint:   managerEndPoint,
		ManagerMetrics:    managerMetrics,
		ManagerToken:      managerToken,
		ManagerClientData: managerClientData,
	}

	instance = container

	return instance
}

func loadRepositoryRequest(config configuration.Configuration) repository.IRepositoryRequest {
	var file repository.IFileManager[action.Request]
	file = repository.NewManagerCsvtFile[action.Request](repository.CSVT_FILE_PATH_REQUEST)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := system.SNAPSHOT_TOPIC_REQUEST
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, action.Request]()
	repository, err := request.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryResponse(config configuration.Configuration) repository.IRepositoryResponse {
	var file repository.IFileManager[action.Response]
	file = repository.NewManagerCsvtFile[action.Response](repository.CSVT_FILE_PATH_RESPONSE)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := system.SNAPSHOT_TOPIC_RESPONSE
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, action.Response]()
	repository, err := response.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryContext(config configuration.Configuration) repository.IRepositoryContext {
	var file repository.IFileManager[dto.DtoContext]
	file = repository.NewManagerCsvtFile[dto.DtoContext](repository.CSVT_FILE_PATH_CONTEXT)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := system.SNAPSHOT_TOPIC_CONTEXT
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, context.Context]()
	repository, err := repository_context.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryCollection(config configuration.Configuration) repository.IRepositoryCollection {
	var file repository.IFileManager[collection_domain.Collection]
	file = repository.NewManagerCsvtFile[collection_domain.Collection](repository.CSVT_FILE_PATH_COLLECTION)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := system.SNAPSHOT_TOPIC_COLLECTION
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, collection_domain.Collection]()
	repository, err := repository_collection.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryGroup(config configuration.Configuration) repository.IRepositoryGroup {
	var file repository.IFileManager[domain.Group]
	file = repository.NewManagerCsvtFile[domain.Group](repository.CSVT_FILE_PATH_GROUP)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := system.SNAPSHOT_TOPIC_GROUP
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, domain.Group]()
	repository, err := repository_group.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryEndPoint(config configuration.Configuration) repository.IRepositoryEndPoint {
	var file repository.IFileManager[mock_domain.EndPoint]
	file = repository.NewManagerCsvtFile[mock_domain.EndPoint](repository.CSVT_FILE_PATH_END_POINT)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := system.SNAPSHOT_TOPIC_END_POINT
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, mock_domain.EndPoint]()
	repository, err := repository_mock.InitializeEndPointRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryMetrics(config configuration.Configuration) repository.IRepositoryMetrics {
	var file repository.IFileManager[mock_domain.Metrics]
	file = repository.NewManagerCsvtFile[mock_domain.Metrics](repository.CSVT_FILE_PATH_METRICS)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := system.SNAPSHOT_TOPIC_END_POINT
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, mock_domain.Metrics]()
	repository, err := repository_mock.InitializeMetricsRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryToken(config configuration.Configuration) repository.IRepositoryToken {
	var file repository.IFileManager[token_domain.Token]
	file = repository.NewManagerCsvtFile[token_domain.Token](repository.CSVT_FILE_PATH_TOKEN)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := system.SNAPSHOT_TOPIC_TOKEN
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, token_domain.Token]()
	repository, err := repository_token.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryClientData(config configuration.Configuration) repository.IRepositoryClientData {
	var file repository.IFileManager[client.ClientData]
	file = repository.NewManagerCsvtFile[client.ClientData](repository.CSVT_FILE_PATH_CLIENT_DATA)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := system.SNAPSHOT_TOPIC_TOKEN
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, client.ClientData]()
	repository, err := repository_client.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadManagerSnapshotFile[T repository.IStructure](topic system.TopicSnapshot, snapshot configuration.Snapshot, file repository.IFileManager[T]) repository.IFileManager[T] {
	return repository.
		BuilderManagerSnapshotFile(topic, file).
		Limit(snapshot.Limit).
		Time(snapshot.Time).
		Make()
}

func loadManagerRequest(
	request repository.IRepositoryRequest,
	response repository.IRepositoryResponse) *repository.ManagerRequest {
	return repository.NewManagerRequest(request, response)
}

func loadManagerContext(context repository.IRepositoryContext) *repository.ManagerContext {
	return repository.NewManagerContext(context)
}

func loadManagerCollection(
	collection repository.IRepositoryCollection,
	managerContext *repository.ManagerContext,
	managerRequest *repository.ManagerRequest) *repository.ManagerCollection {
	return repository.NewManagerCollection(collection, managerContext, managerRequest)
}

func loadManagerHistoric(
	managerRequest *repository.ManagerRequest,
	managerCollection *repository.ManagerCollection) *repository.ManagerHistoric {
	return repository.NewManagerHistoric(managerRequest, managerCollection)
}

func loadManagerGroup(
	group repository.IRepositoryGroup,
	managerCollection *repository.ManagerCollection) *repository.ManagerGroup {
	return repository.NewManagerGroup(group, managerCollection)
}

func loadManagerEndPoint(endPoint repository.IRepositoryEndPoint, metrics *repository.ManagerMetrics) *repository.ManagerEndPoint {
	return repository.NewManagerEndPoint(endPoint, metrics)
}

func loadManagerMetrics(endPoint repository.IRepositoryMetrics) *repository.ManagerMetrics {
	return repository.NewManagerMetrics(endPoint)
}

func loadManagerToken(token repository.IRepositoryToken) *repository.ManagerToken {
	return repository.NewManagerToken(token)
}

func loadManagerClientData(
	client repository.IRepositoryClientData,
	managerCollection *repository.ManagerCollection,
	managerGroup *repository.ManagerGroup) *repository.ManagerClientData {
	return repository.NewManagerClientData(client, managerCollection, managerGroup)
}
