package context

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Rafael24595/go-collections/collection"
)

type DictionaryVariables = collection.Dictionary[string, string]
type DictionaryCategory = collection.Dictionary[string, DictionaryVariables]

type Context struct {
	Id         string             `json:"_id"`
	Status     bool               `json:"status"`
	Timestamp  int64              `json:"timestamp"`
	Dictionary DictionaryCategory `json:"dictionary"`
	Owner      string             `json:"owner"`
	Modified   int64              `json:"modified"`
}

func NewContext(owner string) *Context {
	return &Context{
		Id:         "",
		Status:     true,
		Timestamp:  time.Now().UnixMilli(),
		Dictionary: *collection.DictionaryEmpty[string, DictionaryVariables](),
		Owner:      owner,
		Modified:   time.Now().UnixMilli(),
	}
}

func (c *Context) PutAll(category string, context map[string]string) *Context {
	variables, ok := c.Dictionary.Get(category)
	if !ok {
		c.Dictionary.Put(category, *collection.DictionaryEmpty[string, string]())
		variables, _ = c.Dictionary.Get(category)
	}
	variables.PutAll(context)
	return c
}

func (c Context) Apply(category, source string) string {
	fix := source
	for _, v := range c.IdentifyVariables(category, source) {
		value := ""
		categoryGroup, ok := c.Dictionary.Get(v.Key())
		if ok {
			keyValue, ok := categoryGroup.Get(v.Value())
			if ok {
				value = *keyValue
			}
		}
		fix = strings.ReplaceAll(fix, fmt.Sprintf("${%s}", v.Value()), value)
		fix = strings.ReplaceAll(fix, fmt.Sprintf("${%s.%s}", v.Key(), v.Value()), value)
	}
	return fix
}

func (c Context) IdentifyVariables(category, source string) []collection.Pair[string, string] {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(source, -1)

	results := collection.DictionaryEmpty[string, collection.Pair[string, string]]()
	for _, match := range matches {
		if len(match) == 0 {
			continue
		}

		category := category
		key := match[1]

		fragments := strings.Split(key, ".")
		if len(fragments) > 1 {
			category = fragments[0]
			key = fragments[1]
		}
		
		results.PutIfAbsent(fmt.Sprintf("%s.%s", category, key), collection.NewPair(category, key))
	}

	return results.Values()
}
