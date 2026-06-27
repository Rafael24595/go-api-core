package commons

import (
	"context"
	"maps"
	"os"
	"strings"
	"time"

	topic_snapshot "github.com/Rafael24595/go-api-core/src/commons/system/topic/snapshot"
	domain_session "github.com/Rafael24595/go-api-core/src/domain/session"
	repository_session "github.com/Rafael24595/go-api-core/src/infrastructure/repository/session"

	"github.com/Rafael24595/go-api-core/src/application/session"
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/dependency"
	"github.com/Rafael24595/go-api-core/src/commons/local"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-log/log"
	"github.com/Rafael24595/go-log/log/provider/console"
	"github.com/Rafael24595/go-log/log/provider/file"
	"github.com/Rafael24595/go-log/log/record"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func Initialize(ctx context.Context, kargs map[string]utils.Argument) (*configuration.Configuration, *dependency.DependencyContainer) {
	session := uuid.NewString()
	timestamp := time.Now().UnixMilli()

	store := configLog(ctx, session, kargs)

	mod := ReadGoMod()
	pkg := ReadPackage()
	snp := readSnapshot(kargs)
	config := configuration.Initialize(session, timestamp, kargs, mod, &pkg.Project, snp)

	log.Messagef("Session ID: %s", config.SessionId())
	log.Messagef("Started at: %s", utils.FormatMilliseconds(config.Timestamp()))
	log.Messagef("Dev mode: %v", config.Dev())

	container := dependency.Initialize(config, store)

	repositorySession := loadRepositorySession(config)
	initializeManagerSession(config, repositorySession, container)

	return &config, container
}

func configLog(ctx context.Context, session string, kargs map[string]utils.Argument) *record.Memory {
	store := record.NewMemory()

	instance, ok := kargs["GAC_LOG_INSTANCE"]
	if !ok {
		instance = *utils.ArgumentFrom("")
	}

	var provider log.Provider

	switch strings.ToUpper(instance.String()) {
	case "FILE":
		provider = file.FileProvider{
			Session:     session,
			RecordStore: store,
		}
	default:
		provider = console.ConsoleProvider{
			RecordStore: store,
		}
	}

	if err := log.DefaultFromProvider(ctx, provider); err != nil {
		panic(err.Error())
	}

	return store
}

func loadRepositorySession(config configuration.Configuration) domain_session.RepositorySession {
	var file repository.IFileManager[dto.DtoSession]
	file = repository.NewManagerCsvtFile[dto.DtoSession](repository.CSVT_FILE_PATH_SESSION)

	snapshot := config.Snapshot()
	if snapshot.Enable {
		topic := topic_snapshot.TOPIC_SESSION
		file = loadManagerSnapshotFile(topic, snapshot, file)
	}

	repository, err := repository_session.InitializeRepositoryMemory(file)
	if err != nil {
		local.Panic(err)
	}

	return repository
}

func initializeManagerSession(
	config configuration.Configuration,
	sessions domain_session.RepositorySession,
	container *dependency.DependencyContainer) *session.ManagerSession {
	return session.InitializeManagerSession(config, sessions, container.ManagerSessionData)
}

func loadManagerSnapshotFile[T repository.IStructure](topic topic_snapshot.TopicSnapshot, snapshot configuration.Snapshot, file repository.IFileManager[T]) repository.IFileManager[T] {
	return repository.
		BuilderManagerSnapshotFile(topic, file).
		Limit(snapshot.Limit).
		Time(snapshot.Time).
		Make()
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
		local.Panic(err)
	}

	result := configuration.DecodeMod(file)
	err = file.Close()
	if err != nil {
		local.Panic(err)
	}

	return result
}

func ReadPackage() *configuration.Package {
	file, err := os.Open("go.package.yml")
	if err != nil {
		local.Panicf("Error opening go.package.yml: %v", err)
	}

	var pkg configuration.Package
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&pkg); err != nil {
		local.Panicf("Error decoding YAML: %v", err)
	}

	if err := file.Close(); err != nil {
		local.Panicf("Error closing file: %v", err)
	}

	return &pkg
}

func readSnapshot(kargs map[string]utils.Argument) *configuration.Snapshot {
	enable := kargs["GAC_ENABLE_SNAPSHOT"].Boold(false)
	time := kargs["GAC_SNAPSHOT_TIME"].Int64d(0)
	limit := kargs["GAC_SNAPSHOT_LIMIT"].Intd(1)

	return &configuration.Snapshot{
		Enable: enable,
		Time:   time * int64(readSnapshotUnit(kargs)),
		Limit:  limit,
	}
}

func readSnapshotUnit(kargs map[string]utils.Argument) time.Duration {
	switch kargs["GAC_SNAPSHOT_UNIT"].String() {
	case "MILLISECOND":
		return time.Millisecond
	case "SECOND":
		return time.Second
	case "MINUTE":
		return time.Minute
	case "HOUR":
		return time.Hour
	case "DAY":
		return time.Hour * 24
	case "WEEK":
		return time.Hour * 24 * 7
	default:
		return time.Hour
	}
}
