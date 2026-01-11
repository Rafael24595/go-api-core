package system

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/format"
	"github.com/Rafael24595/go-collections/collection"
)

type TopicSnapshot string

type SnapshotMeta struct {
	isCore      bool
	Description string
	CsvPath     string
}

type SnapshotExtension struct {
	Topic       TopicSnapshot
	Description string
	CsvPath     string
}

const (
	SNAPSHOT_TOPIC_CONTEXT     TopicSnapshot = "snpsh_ctx"
	SNAPSHOT_TOPIC_REQUEST     TopicSnapshot = "snpsh_rqt"
	SNAPSHOT_TOPIC_RESPONSE    TopicSnapshot = "snpsh_rsp"
	SNAPSHOT_TOPIC_COLLECTION  TopicSnapshot = "snpsh_coll"
	SNAPSHOT_TOPIC_GROUP       TopicSnapshot = "snpsh_grp"
	SNAPSHOT_TOPIC_END_POINT   TopicSnapshot = "snpsh_ept"
	SNAPSHOT_TOPIC_METRICS     TopicSnapshot = "snpsh_epm"
	SNAPSHOT_TOPIC_TOKEN       TopicSnapshot = "snpsh_tkn"
	SNAPSHOT_TOPIC_SESSION     TopicSnapshot = "snpsh_ses"
	SNAPSHOT_TOPIC_CLIENT_DATA TopicSnapshot = "snpsh_cld"
)

var snapshotMeta = map[TopicSnapshot]SnapshotMeta{
	SNAPSHOT_TOPIC_CONTEXT: {
		isCore:      true,
		Description: "Represents a snapshot of contextual data.",
		CsvPath:     "./db/snapshot/context",
	},
	SNAPSHOT_TOPIC_REQUEST: {
		isCore:      true,
		Description: "Represents a snapshot of request data.",
		CsvPath:     "./db/snapshot/request",
	},
	SNAPSHOT_TOPIC_RESPONSE: {
		isCore:      true,
		Description: "Represents a snapshot of response data.",
		CsvPath:     "./db/snapshot/response",
	},
	SNAPSHOT_TOPIC_COLLECTION: {
		isCore:      true,
		Description: "Represents a snapshot of collection data.",
		CsvPath:     "./db/snapshot/collection",
	},
	SNAPSHOT_TOPIC_GROUP: {
		isCore:      true,
		Description: "Represents a snapshot of group data.",
		CsvPath:     "./db/snapshot/group",
	},
	SNAPSHOT_TOPIC_END_POINT: {
		isCore:      true,
		Description: "Represents a snapshot of mocked API endpoint data.",
		CsvPath:     "./db/snapshot/end_point",
	},
	SNAPSHOT_TOPIC_METRICS: {
		isCore:      true,
		Description: "Represents a snapshot of mocked API endpoint metrics.",
		CsvPath:     "./db/snapshot/metrics",
	},
	SNAPSHOT_TOPIC_TOKEN: {
		isCore:      true,
		Description: "Represents a snapshot of user token data.",
		CsvPath:     "./db/snapshot/token",
	},
	SNAPSHOT_TOPIC_SESSION: {
		isCore:      true,
		Description: "Represents a snapshot of user session data.",
		CsvPath:     "./db/snapshot/session",
	},
	SNAPSHOT_TOPIC_CLIENT_DATA: {
		isCore:      true,
		Description: "Represents a snapshot of user client data.",
		CsvPath:     "./db/snapshot/client_data",
	},
}

const CSVT_SNAPSHOT_PATH_MISC string = "./db/snapshot/misc"

func allTopicSnapshots() []TopicSnapshot {
	keys := make([]TopicSnapshot, 0, len(snapshotMeta))
	for ts := range snapshotMeta {
		keys = append(keys, ts)
	}
	return keys
}

func ExtendMany(topics ...SnapshotExtension) []TopicSnapshot {
	result := make([]TopicSnapshot, 0)
	for _, t := range topics {
		r, ok := Extend(t)
		if ok {
			result = append(result, r)
		}
	}
	return result
}

func Extend(topic SnapshotExtension) (TopicSnapshot, bool) {
	old, ok := snapshotMeta[topic.Topic]
	if ok && old.isCore {
		return topic.Topic, false
	}

	snapshotMeta[topic.Topic] = SnapshotMeta{
		isCore:      false,
		Description: topic.Description,
		CsvPath:     topic.CsvPath,
	}
	return topic.Topic, true
}

func TopicSnapshotFromString(s string) (TopicSnapshot, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, t := range allTopicSnapshots() {
		if string(t) == s {
			return t, true
		}
	}
	return "", false
}

func FilterTopicSnapshot(topics []string) []TopicSnapshot {
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

func (t TopicSnapshot) Description() string {
	if meta, ok := snapshotMeta[t]; ok {
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
	if meta, ok := snapshotMeta[t]; ok {
		return meta.CsvPath
	}
	return CSVT_SNAPSHOT_PATH_MISC
}

func (t TopicSnapshot) TopicSnapshotSaveInput() string {
	return fmt.Sprintf("%s_sav_inp", string(t))
}

func (t TopicSnapshot) TopicSnapshotAppyInput() string {
	return fmt.Sprintf("%s_apl_inp", string(t))
}

func (t TopicSnapshot) TopicSnapshotApplyOutput() string {
	return fmt.Sprintf("%s_apl_out", string(t))
}

func (t TopicSnapshot) TopicSnapshotRemoveInput() string {
	return fmt.Sprintf("%s_rmv_inp", string(t))
}
