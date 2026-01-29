package curl

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	auth_strategy "github.com/Rafael24595/go-api-core/src/domain/action/auth/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	body_strategy "github.com/Rafael24595/go-api-core/src/domain/action/body/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
	"github.com/Rafael24595/go-collections/collection"
)

const (
	WINDOWS_NEW_LINE  = "\r\n"
	UNIX_CONTINUATION = "\\\n"
	PWSH_CONTINUATION = "`\n"
	CMD_CONTINUATION  = "^\n"
)

type curlData struct {
	uri    string
	tuples []utils.CmdTuple
}

func Unmarshal(curl []byte) (*action.Request, error) {
	data, err := decode(curl)
	if err != nil {
		return nil, err
	}

	uri, query := processUri(data.uri)

	name := fmt.Sprintf("[cURL] %s", uri)
	request := action.NewRequest(name, domain.GET, uri)
	request.Query = *query

	for _, v := range data.tuples {
		switch v.Flag {
		case "-X":
			request.Method = domain.HttpMethod(v.Data)
		case "-H", "--header":
			request = processHeader(v.Data, request)
		case "-d", "--data", "--data-raw":
			request = processDocument(v.Data, request)
		case "--data-binary":
			request = processBinary(v.Data, request)
		case "-F", "--form":
			request = processFormData(v.Data, request)
		case "-u", "--user":
			request = processBasicAuth(v.Data, request)
		case "-b", "--cookie":
			request = processCookie(v.Data, request)
		}
	}

	return request, nil
}

func processUri(uri string) (string, *query.Queries) {
	queries := query.NewQueries()

	fragments := strings.SplitN(uri, "?", 2)
	if len(fragments) == 1 {
		return uri, queries
	}

	values, err := url.ParseQuery(fragments[1])
	if err != nil {
		return uri, queries
	}

	for key, vals := range values {
		for _, v := range vals {
			queries.Add(key, v)
		}
	}

	return fragments[0], queries
}

func processDocument(data string, request *action.Request) *action.Request {
	request.Body = *body_strategy.DocumentBody(true, domain.Text, data)
	if request.Method == domain.GET {
		request.Method = domain.POST
	}

	return request
}

func processFormData(data string, request *action.Request) *action.Request {
	parts := strings.SplitN(data, "=", 2)
	key := strings.TrimSpace(parts[0])

	value := ""
	if len(parts) > 1 {
		value = strings.TrimSpace(parts[1])
	}

	var parameter *body.BodyParameter
	if path, ok := strings.CutPrefix(value, "@"); ok {
		ext := strings.TrimPrefix(filepath.Ext(path), ".")
		parameter = body.NewFileParameterActive(ext, "", path)
	} else {
		parameter = body.NewParameterActive(value)
	}

	body := body_strategy.AddFormData(&request.Body, key, parameter)

	request.Body = *body

	if request.Method == domain.GET {
		request.Method = domain.POST
	}

	return request
}

func processBinary(data string, request *action.Request) *action.Request {
	value := strings.TrimSpace(data)
	if path, ok := strings.CutPrefix(value, "@"); ok {
		value = path
	}

	return processDocument(value, request)
}

func processHeader(data string, request *action.Request) *action.Request {
	for hs := range strings.SplitSeq(data, ";") {
		parts := strings.SplitN(hs, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			request.Header.Add(key, value)
		}
	}
	return request
}

func processBasicAuth(data string, request *action.Request) *action.Request {
	parts := strings.SplitN(data, ":", 2)
	user := strings.TrimSpace(parts[0])

	pass := ""
	if len(parts) > 1 {
		pass = strings.TrimSpace(parts[1])
	}

	auth := auth_strategy.BasicAuth(true, user, pass)
	request.Auth.PutAuth(*auth)
	request.Auth.Status = true
	return request
}

func processCookie(data string, request *action.Request) *action.Request {
	for hs := range strings.SplitSeq(data, ";") {
		parts := strings.SplitN(hs, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			request.Cookie.Put(key, value)
		}
	}
	return request
}

func decode(curl []byte) (*curlData, error) {
	clean := clean(curl)

	args, err := utils.SplitCommand(clean)
	if err != nil {
		return nil, err
	}

	uri := ""
	tuples := make([]utils.CmdTuple, 0)

	fragments := collection.VectorFromList(args)
	head, ok := fragments.Shift()
	if !ok || head != "curl" {
		return nil, errors.New("the command is not a valid curl sentence")
	}

	for fragments.Size() > 0 {
		flag, ok := fragments.Shift()
		if !ok {
			return nil, errors.New("the command flag could not be empty")
		}

		url, ok, err := cutUri(flag, fragments)
		if err != nil {
			return nil, err
		}

		if ok {
			uri = url
			continue
		}

		data, ok := fragments.Shift()
		if !ok {
			return nil, errors.New("the command flag data could not be empty")
		}

		tuples = append(tuples, utils.CmdTuple{
			Flag: flag,
			Data: data,
		})
	}

	return &curlData{
		uri:    uri,
		tuples: tuples,
	}, nil
}

func clean(curl []byte) string {
	clean := strings.ReplaceAll(string(curl), WINDOWS_NEW_LINE, "\n")
	clean = strings.ReplaceAll(clean, UNIX_CONTINUATION, " ")
	clean = strings.ReplaceAll(clean, PWSH_CONTINUATION, " ")
	clean = strings.ReplaceAll(clean, CMD_CONTINUATION, " ")
	return strings.TrimSpace(clean)
}

func cutUri(flag string, fragmensts *collection.Vector[string]) (string, bool, error) {
	if flag == "--uri" || flag == "--url" {
		uri, ok := fragmensts.Shift()
		if !ok {
			return "", false, errors.New("the uri is not defined")
		}

		flag = uri
	} else if strings.HasPrefix(flag, "-") &&
		(!strings.Contains(flag, "://") && !strings.HasPrefix(flag, "/")) {
		return "", false, nil
	}

	u, err := url.Parse(flag)
	if err != nil {
		return "", false, err
	}

	if u.Scheme == "" && !strings.HasPrefix(flag, "/") {
		return "", false, fmt.Errorf("invalid URI: %s", flag)
	}

	return flag, true, nil
}
