package topic_repository

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	"github.com/Rafael24595/go-collections/collection"
)

type TopicRepository string

type TopicMeta struct {
	isCore      bool
	Description string
}

type Extension struct {
	Topic       TopicRepository
	Description string
}

const (
	TOPIC_CONTEXT     TopicRepository = "rep_ctx"
	TOPIC_REQUEST     TopicRepository = "rep_rqt"
	TOPIC_RESPONSE    TopicRepository = "rep_rsp"
	TOPIC_COLLECTION  TopicRepository = "rep_coll"
	TOPIC_GROUP       TopicRepository = "rep_grp"
	TOPIC_END_POINT   TopicRepository = "rep_ept"
	TOPIC_METRICS     TopicRepository = "rep_epm"
	TOPIC_TOKEN       TopicRepository = "rep_tkn"
	TOPIC_SESSION     TopicRepository = "rep_ses"
	TOPIC_CLIENT_DATA TopicRepository = "rep_cld"
)

var snapshotMeta = map[TopicRepository]TopicMeta{
	TOPIC_CONTEXT: {
		isCore:      true,
		Description: "Represents the repository of contextual data.",
	},
	TOPIC_REQUEST: {
		isCore:      true,
		Description: "Represents the repository of request data.",
	},
	TOPIC_RESPONSE: {
		isCore:      true,
		Description: "Represents the repository of response data.",
	},
	TOPIC_COLLECTION: {
		isCore:      true,
		Description: "Represents the repository of collection data.",
	},
	TOPIC_GROUP: {
		isCore:      true,
		Description: "Represents the repository of group data.",
	},
	TOPIC_END_POINT: {
		isCore:      true,
		Description: "Represents the repository of mocked API endpoint data.",
	},
	TOPIC_METRICS: {
		isCore:      true,
		Description: "Represents the repository of mocked API endpoint metrics.",
	},
	TOPIC_TOKEN: {
		isCore:      true,
		Description: "Represents the repository of user token data.",
	},
	TOPIC_SESSION: {
		isCore:      true,
		Description: "Represents the repository of user session data.",
	},
	TOPIC_CLIENT_DATA: {
		isCore:      true,
		Description: "Represents the repository of user client data.",
	},
}

func allTopicRepositorys() []TopicRepository {
	keys := make([]TopicRepository, 0, len(snapshotMeta))
	for ts := range snapshotMeta {
		keys = append(keys, ts)
	}
	return keys
}

func ExtendMany(topics ...Extension) []TopicRepository {
	result := make([]TopicRepository, 0)
	for _, t := range topics {
		r, ok := Extend(t)
		if ok {
			result = append(result, r)
		}
	}
	return result
}

func Extend(topic Extension) (TopicRepository, bool) {
	old, ok := snapshotMeta[topic.Topic]
	if ok && old.isCore {
		return topic.Topic, false
	}

	snapshotMeta[topic.Topic] = TopicMeta{
		isCore:      false,
		Description: topic.Description,
	}

	return topic.Topic, true
}

func TopicFromString(s string) (TopicRepository, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, t := range allTopicRepositorys() {
		if string(t) == s {
			return t, true
		}
	}
	return "", false
}

func FindTopics(topics []string) []TopicRepository {
	cache := make(map[TopicRepository]byte)
	for _, c := range topics {
		for _, t := range allTopicRepositorys() {
			if strings.HasPrefix(c, string(t)) {
				cache[t] = byte(0)
			}
		}
	}
	return collection.DictionaryFromMap(cache).Keys()
}

func (t TopicRepository) Meta() TopicMeta {
	return snapshotMeta[t]
}

func (t TopicRepository) Description() string {
	if meta, ok := snapshotMeta[t]; ok {
		return meta.Description
	}
	return "Unknown topic repository type"
}

func (t TopicRepository) FindAcction(action string) (*topic.TopicAction, bool) {
	reload := t.ActionReload()

	switch action {
	case reload.Code:
		return &reload, true
	}

	return nil, false
}

func (t TopicRepository) ActionReload() topic.TopicAction {
	return topic.TopicAction{
		Parent:      string(t),
		Code:        fmt.Sprintf("%s_rel", string(t)),
		Description: "Reloads the repository",
	}
}
