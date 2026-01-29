package dependency

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/application/manager"
	"github.com/Rafael24595/go-api-core/src/application/session"
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	topic_snapshot "github.com/Rafael24595/go-api-core/src/commons/system/topic/snapshot"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	collection_domain "github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/domain/group"
	domain_mock "github.com/Rafael24595/go-api-core/src/domain/mock"
	domain_session "github.com/Rafael24595/go-api-core/src/domain/session"
	domain_token "github.com/Rafael24595/go-api-core/src/domain/token"
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

var (
	instance *DependencyContainer
	once     sync.Once
)

type DependencyContainer struct {
	RepositoryContext  context.Repository
	ManagerRequest     *manager.ManagerRequest
	ManagerContext     *manager.ManagerContext
	ManagerCollection  *manager.ManagerCollection
	ManagerHistoric    *manager.ManagerHistoric
	ManagerGroup       *manager.ManagerGroup
	ManagerEndPoint    *manager.ManagerEndPoint
	ManagerMetrics     *manager.ManagerMetrics
	ManagerToken       *manager.ManagerToken
	ManagerSessionData *session.ManagerSessionData
}

func Initialize(config configuration.Configuration) *DependencyContainer {
	once.Do(func() {
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
		managerSessionData := loadManagerSessionData(repositoryClient, managerCollection, managerGroup)

		container := &DependencyContainer{
			RepositoryContext:  repositoryContext,
			ManagerRequest:     managerRequest,
			ManagerContext:     managerContext,
			ManagerCollection:  managerCollection,
			ManagerHistoric:    managerHistoric,
			ManagerGroup:       managerGroup,
			ManagerEndPoint:    managerEndPoint,
			ManagerMetrics:     managerMetrics,
			ManagerToken:       managerToken,
			ManagerSessionData: managerSessionData,
		}

		instance = container
	})

	return instance
}

func Instance() *DependencyContainer {
	if instance == nil {
		panic("depencency container is not instanced")
	}

	return instance
}

func loadRepositoryRequest(config configuration.Configuration) action.RepositoryRequest {
	var file repository.IFileManager[action.Request]
	file = repository.NewManagerCsvtFile[action.Request](repository.CSVT_FILE_PATH_REQUEST)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_REQUEST
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, action.Request]()
	repository, err := request.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryResponse(config configuration.Configuration) action.RepositoryResponse {
	var file repository.IFileManager[action.Response]
	file = repository.NewManagerCsvtFile[action.Response](repository.CSVT_FILE_PATH_RESPONSE)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_RESPONSE
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, action.Response]()
	repository, err := response.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryContext(config configuration.Configuration) context.Repository {
	var file repository.IFileManager[dto.DtoContext]
	file = repository.NewManagerCsvtFile[dto.DtoContext](repository.CSVT_FILE_PATH_CONTEXT)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_CONTEXT
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, context.Context]()
	repository, err := repository_context.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryCollection(config configuration.Configuration) collection_domain.Repository {
	var file repository.IFileManager[collection_domain.Collection]
	file = repository.NewManagerCsvtFile[collection_domain.Collection](repository.CSVT_FILE_PATH_COLLECTION)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_COLLECTION
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, collection_domain.Collection]()
	repository, err := repository_collection.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryGroup(config configuration.Configuration) group.Repository {
	var file repository.IFileManager[group.Group]
	file = repository.NewManagerCsvtFile[group.Group](repository.CSVT_FILE_PATH_GROUP)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_GROUP
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, group.Group]()
	repository, err := repository_group.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryEndPoint(config configuration.Configuration) domain_mock.RepositoryEndPoint {
	var file repository.IFileManager[domain_mock.EndPoint]
	file = repository.NewManagerCsvtFile[domain_mock.EndPoint](repository.CSVT_FILE_PATH_END_POINT)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_END_POINT
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, domain_mock.EndPoint]()
	repository, err := repository_mock.InitializeEndPointRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryMetrics(config configuration.Configuration) domain_mock.RepositoryMetrics {
	var file repository.IFileManager[domain_mock.Metrics]
	file = repository.NewManagerCsvtFile[domain_mock.Metrics](repository.CSVT_FILE_PATH_METRICS)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_END_POINT
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, domain_mock.Metrics]()
	repository, err := repository_mock.InitializeMetricsRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryToken(config configuration.Configuration) domain_token.Repository {
	var file repository.IFileManager[domain_token.Token]
	file = repository.NewManagerCsvtFile[domain_token.Token](repository.CSVT_FILE_PATH_TOKEN)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_TOKEN
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, domain_token.Token]()
	repository, err := repository_token.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadRepositoryClientData(config configuration.Configuration) domain_session.RepositorySessionData {
	var file repository.IFileManager[domain_session.ClientData]
	file = repository.NewManagerCsvtFile[domain_session.ClientData](repository.CSVT_FILE_PATH_CLIENT_DATA)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_TOKEN
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	impl := collection.DictionarySyncEmpty[string, domain_session.ClientData]()
	repository, err := repository_client.InitializeRepositoryMemory(impl, file)
	if err != nil {
		log.Panic(err)
	}

	return repository
}

func loadManagerSnapshotFile[T repository.IStructure](
	topic topic_snapshot.TopicSnapshot,
	snapshot configuration.Snapshot,
	file repository.IFileManager[T]) repository.IFileManager[T] {
	return repository.BuilderManagerSnapshotFile(topic, file).
		Limit(snapshot.Limit).
		Time(snapshot.Time).
		Make()
}

func loadManagerRequest(
	request action.RepositoryRequest,
	response action.RepositoryResponse) *manager.ManagerRequest {
	return manager.NewManagerRequest(request, response)
}

func loadManagerContext(context context.Repository) *manager.ManagerContext {
	return manager.NewManagerContext(context)
}

func loadManagerCollection(
	collection collection_domain.Repository,
	managerContext *manager.ManagerContext,
	managerRequest *manager.ManagerRequest) *manager.ManagerCollection {
	return manager.NewManagerCollection(collection, managerContext, managerRequest)
}

func loadManagerHistoric(
	managerRequest *manager.ManagerRequest,
	managerCollection *manager.ManagerCollection) *manager.ManagerHistoric {
	return manager.NewManagerHistoric(managerRequest, managerCollection)
}

func loadManagerGroup(
	group group.Repository,
	managerCollection *manager.ManagerCollection) *manager.ManagerGroup {
	return manager.NewManagerGroup(group, managerCollection)
}

func loadManagerEndPoint(endPoint domain_mock.RepositoryEndPoint, metrics *manager.ManagerMetrics) *manager.ManagerEndPoint {
	return manager.NewManagerEndPoint(endPoint, metrics)
}

func loadManagerMetrics(endPoint domain_mock.RepositoryMetrics) *manager.ManagerMetrics {
	return manager.NewManagerMetrics(endPoint)
}

func loadManagerToken(token domain_token.Repository) *manager.ManagerToken {
	return manager.NewManagerToken(token)
}

func loadManagerSessionData(
	client domain_session.RepositorySessionData,
	managerCollection *manager.ManagerCollection,
	managerGroup *manager.ManagerGroup) *session.ManagerSessionData {
	return session.NewManagerSessionData(client, managerCollection, managerGroup)
}
