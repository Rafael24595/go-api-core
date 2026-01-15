package topic_snapshot

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/format"
	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	"github.com/Rafael24595/go-collections/collection"
)

type TopicSnapshot string

type TopicMeta struct {
	isCore      bool
	Description string
	CsvPath     string
	Repository  topic_repository.TopicRepository
}

type Extension struct {
	Topic       TopicSnapshot
	Description string
	CsvPath     string
	Repository  topic_repository.TopicRepository
}

const (
	TOPIC_CONTEXT     TopicSnapshot = "snpsh_ctx"
	TOPIC_REQUEST     TopicSnapshot = "snpsh_rqt"
	TOPIC_RESPONSE    TopicSnapshot = "snpsh_rsp"
	TOPIC_COLLECTION  TopicSnapshot = "snpsh_coll"
	TOPIC_GROUP       TopicSnapshot = "snpsh_grp"
	TOPIC_END_POINT   TopicSnapshot = "snpsh_ept"
	TOPIC_METRICS     TopicSnapshot = "snpsh_epm"
	TOPIC_TOKEN       TopicSnapshot = "snpsh_tkn"
	TOPIC_SESSION     TopicSnapshot = "snpsh_ses"
	TOPIC_CLIENT_DATA TopicSnapshot = "snpsh_cld"
)

var meta = map[TopicSnapshot]TopicMeta{
	TOPIC_CONTEXT: {
		isCore:      true,
		Description: "Represents a snapshot of contextual data.",
		CsvPath:     "./db/snapshot/context",
		Repository:  topic_repository.TOPIC_CONTEXT,
	},
	TOPIC_REQUEST: {
		isCore:      true,
		Description: "Represents a snapshot of request data.",
		CsvPath:     "./db/snapshot/request",
		Repository:  topic_repository.TOPIC_REQUEST,
	},
	TOPIC_RESPONSE: {
		isCore:      true,
		Description: "Represents a snapshot of response data.",
		CsvPath:     "./db/snapshot/response",
		Repository:  topic_repository.TOPIC_RESPONSE,
	},
	TOPIC_COLLECTION: {
		isCore:      true,
		Description: "Represents a snapshot of collection data.",
		CsvPath:     "./db/snapshot/collection",
		Repository:  topic_repository.TOPIC_COLLECTION,
	},
	TOPIC_GROUP: {
		isCore:      true,
		Description: "Represents a snapshot of group data.",
		CsvPath:     "./db/snapshot/group",
		Repository:  topic_repository.TOPIC_GROUP,
	},
	TOPIC_END_POINT: {
		isCore:      true,
		Description: "Represents a snapshot of mocked API endpoint data.",
		CsvPath:     "./db/snapshot/end_point",
		Repository:  topic_repository.TOPIC_END_POINT,
	},
	TOPIC_METRICS: {
		isCore:      true,
		Description: "Represents a snapshot of mocked API endpoint metrics.",
		CsvPath:     "./db/snapshot/metrics",
		Repository:  topic_repository.TOPIC_METRICS,
	},
	TOPIC_TOKEN: {
		isCore:      true,
		Description: "Represents a snapshot of user token data.",
		CsvPath:     "./db/snapshot/token",
		Repository:  topic_repository.TOPIC_TOKEN,
	},
	TOPIC_SESSION: {
		isCore:      true,
		Description: "Represents a snapshot of user session data.",
		CsvPath:     "./db/snapshot/session",
		Repository:  topic_repository.TOPIC_SESSION,
	},
	TOPIC_CLIENT_DATA: {
		isCore:      true,
		Description: "Represents a snapshot of user client data.",
		CsvPath:     "./db/snapshot/client_data",
		Repository:  topic_repository.TOPIC_CLIENT_DATA,
	},
}

const CSVT_PATH_MISC string = "./db/snapshot/misc"

func allTopicSnapshots() []TopicSnapshot {
	keys := make([]TopicSnapshot, 0, len(meta))
	for ts := range meta {
		keys = append(keys, ts)
	}
	return keys
}

func ExtendMany(topics ...Extension) []TopicSnapshot {
	result := make([]TopicSnapshot, 0)
	for _, t := range topics {
		r, ok := Extend(t)
		if ok {
			result = append(result, r)
		}
	}
	return result
}

func Extend(topic Extension) (TopicSnapshot, bool) {
	old, ok := meta[topic.Topic]
	if ok && old.isCore {
		return topic.Topic, false
	}

	meta[topic.Topic] = TopicMeta{
		isCore:      false,
		Description: topic.Description,
		CsvPath:     topic.CsvPath,
		Repository:  topic.Repository,
	}
	return topic.Topic, true
}

func TopicFromString(s string) (TopicSnapshot, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, t := range allTopicSnapshots() {
		if string(t) == s {
			return t, true
		}
	}
	return "", false
}

func FindTopics(topics []string) []TopicSnapshot {
	cache := make(map[TopicSnapshot]byte)
	for _, c := range topics {
		for _, t := range allTopicSnapshots() {
			if strings.HasPrefix(c, string(t)) {
				cache[t] = byte(0)
			}
		}
	}
	return collection.DictionaryFromMap(cache).Keys()
}

func (t TopicSnapshot) Meta() TopicMeta {
	return meta[t]
}

func (t TopicSnapshot) Description() string {
	if meta, ok := meta[t]; ok {
		return meta.Description
	}
	return "Unknown topic snapshot type"
}

func (t TopicSnapshot) Path(frmt format.DataFormat) (string, bool) {
	switch frmt {
	case format.CSVT:
		return t.CsvtPath(), true
	}
	return "", false
}

func (t TopicSnapshot) CsvtPath() string {
	if meta, ok := meta[t]; ok {
		return meta.CsvPath
	}
	return CSVT_PATH_MISC
}

func (t TopicSnapshot) FindAcction(action string) (*topic.TopicAction, bool) {
	save := t.ActionSave()
	apply := t.ActionAppy()
	remove := t.ActionRemove()

	switch action {
	case save.Code:
		return &save, true
	case apply.Code:
		return &apply, true
	case remove.Code:
		return &remove, true
	}

	return nil, false
}

func (t TopicSnapshot) ActionSave() topic.TopicAction {
	return topic.TopicAction{
		Parent:      string(t),
		Code:        fmt.Sprintf("%s_sav_inp", string(t)),
		Description: "Creates an snapshot from curren status",
	}
}

func (t TopicSnapshot) ActionAppy() topic.TopicAction {
	return topic.TopicAction{
		Parent:      string(t),
		Code:        fmt.Sprintf("%s_apl_inp", string(t)),
		Description: "Applies an specific snapshot and reloads the repository",
	}
}

func (t TopicSnapshot) ActionRemove() topic.TopicAction {
	return topic.TopicAction{
		Parent:      string(t),
		Code:        fmt.Sprintf("%s_rmv_inp", string(t)),
		Description: "Removes an specific snapshot",
	}
}
