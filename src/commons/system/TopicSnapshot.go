package system

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/format"
	"github.com/Rafael24595/go-collections/collection"
)

type TopicSnapshot string

const (
	SNAPSHOT_TOPIC_CONTEXT    TopicSnapshot = "snpsh_ctx"
	SNAPSHOT_TOPIC_REQUEST    TopicSnapshot = "snpsh_rqt"
	SNAPSHOT_TOPIC_RESPONSE   TopicSnapshot = "snpsh_rsp"
	SNAPSHOT_TOPIC_COLLECTION TopicSnapshot = "snpsh_coll"
	SNAPSHOT_TOPIC_GROUP      TopicSnapshot = "snpsh_grp"
	SNAPSHOT_TOPIC_END_POINT  TopicSnapshot = "snpsh_ept"
	SNAPSHOT_TOPIC_TOKEN      TopicSnapshot = "snpsh_tkn"
	SNAPSHOT_TOPIC_SESSION    TopicSnapshot = "snpsh_ses"
)

const (
	CSVT_SNAPSHOT_PATH_CONTEXT    string = "./db/snapshot/context"
	CSVT_SNAPSHOT_PATH_REQUEST    string = "./db/snapshot/request"
	CSVT_SNAPSHOT_PATH_RESPONSE   string = "./db/snapshot/response"
	CSVT_SNAPSHOT_PATH_COLLECTION string = "./db/snapshot/collection"
	CSVT_SNAPSHOT_PATH_GROUP      string = "./db/snapshot/group"
	CSVT_SNAPSHOT_PATH_END_POINT  string = "./db/snapshot/end_point"
	CSVT_SNAPSHOT_PATH_TOKEN      string = "./db/snapshot/token"
	CSVT_SNAPSHOT_PATH_SESSION    string = "./db/snapshot/session"
	CSVT_SNAPSHOT_PATH_MISC       string = "./db/snapshot/misc"
)

var allTopicSnapshots = []TopicSnapshot{
	SNAPSHOT_TOPIC_CONTEXT,
	SNAPSHOT_TOPIC_REQUEST,
	SNAPSHOT_TOPIC_RESPONSE,
	SNAPSHOT_TOPIC_COLLECTION,
	SNAPSHOT_TOPIC_GROUP,
	SNAPSHOT_TOPIC_END_POINT,
	SNAPSHOT_TOPIC_TOKEN,
	SNAPSHOT_TOPIC_SESSION,
}

var topicDescriptions = map[TopicSnapshot]string{
	SNAPSHOT_TOPIC_CONTEXT:    "Represents a snapshot of contextual data.",
	SNAPSHOT_TOPIC_REQUEST:    "Represents a snapshot of request data.",
	SNAPSHOT_TOPIC_RESPONSE:   "Represents a snapshot of response data.",
	SNAPSHOT_TOPIC_COLLECTION: "Represents a snapshot of collection data.",
	SNAPSHOT_TOPIC_GROUP:      "Represents a snapshot of group data.",
	SNAPSHOT_TOPIC_END_POINT:  "Represents a snapshot of mocked API endpoint data.",
	SNAPSHOT_TOPIC_TOKEN:      "Represents a snapshot of user token data.",
	SNAPSHOT_TOPIC_SESSION:    "Represents a snapshot of user session data.",
}

var topicCsvPath = map[TopicSnapshot]string{
	SNAPSHOT_TOPIC_CONTEXT:    CSVT_SNAPSHOT_PATH_CONTEXT,
	SNAPSHOT_TOPIC_REQUEST:    CSVT_SNAPSHOT_PATH_REQUEST,
	SNAPSHOT_TOPIC_RESPONSE:   CSVT_SNAPSHOT_PATH_RESPONSE,
	SNAPSHOT_TOPIC_COLLECTION: CSVT_SNAPSHOT_PATH_COLLECTION,
	SNAPSHOT_TOPIC_GROUP:      CSVT_SNAPSHOT_PATH_GROUP,
	SNAPSHOT_TOPIC_END_POINT:  CSVT_SNAPSHOT_PATH_END_POINT,
	SNAPSHOT_TOPIC_TOKEN:      CSVT_SNAPSHOT_PATH_TOKEN,
	SNAPSHOT_TOPIC_SESSION:    CSVT_SNAPSHOT_PATH_SESSION,
}

func TopicSnapshotFromString(s string) (TopicSnapshot, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, t := range allTopicSnapshots {
		if string(t) == s {
			return t, true
		}
	}
	return "", false
}

func FilterTopicSnapshot(topics []string) []TopicSnapshot {
	cache := make(map[TopicSnapshot]byte)
	for _, c := range topics {
		for _, t := range allTopicSnapshots {
			if strings.HasPrefix(c, string(t)) {
				cache[t] = byte(0)
			}
		}
	}
	return collection.DictionaryFromMap(cache).Keys()
}

func (t TopicSnapshot) Description() string {
	if desc, ok := topicDescriptions[t]; ok {
		return desc
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
	if desc, ok := topicCsvPath[t]; ok {
		return desc
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
