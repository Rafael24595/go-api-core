package formatter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/domain/action"
	auth_strategy "github.com/Rafael24595/go-api-core/src/domain/action/auth/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	body_strategy "github.com/Rafael24595/go-api-core/src/domain/action/body/strategy"

	"github.com/Rafael24595/go-api-core/src/domain/context"
)

func ToCurlWithContext(ctx *context.Context, req *action.Request, inline bool) (string, error) {
	req = context.ProcessRequest(req, ctx)
	return ToCurl(req, inline)
}

func ToCurl(req *action.Request, inline bool) (string, error) {
	buffer := make([]string, 0)

	method := strings.ToUpper(req.Method.String())
	url := strings.TrimSpace(req.Uri)

	if method == "" || url == "" {
		return "", errors.New("the method or the URI are empty")
	}

	req = auth_strategy.ApplyAuth(req)

	query := queryToCurl(req)

	buffer = append(buffer, fmt.Sprintf("curl -X %s %s%s", method, url, query))

	cookies := cookiesToCurl(req)
	buffer = append(buffer, cookies...)

	headers := headersToCurl(req)
	buffer = append(buffer, headers...)

	payload := bodyToCurl(req)
	buffer = append(buffer, payload...)

	for i := 0; i < len(buffer); i++ {
		buffer[i] = strings.ReplaceAll(buffer[i], "\n", " ")
	}

	delimiter := " \\\n "
	if inline {
		delimiter = " "
	}

	return strings.Join(buffer, delimiter), nil
}

func queryToCurl(req *action.Request) string {
	buffer := make([]string, 0)
	for k, q := range req.Query.Queries {
		values := make([]string, 0)
		for _, v := range q {
			if !v.Status {
				continue
			}

			value := strings.TrimSpace(v.Value)
			values = append(values, value)
		}

		if len(values) > 0 {
			result := strings.Join(values, ",")
			buffer = append(buffer, fmt.Sprintf("%s=%s", k, result))
		}
	}

	if len(buffer) == 0 {
		return ""
	}

	return fmt.Sprintf("?%s", strings.Join(buffer, "&"))
}

func cookiesToCurl(req *action.Request) []string {
	buffer := make([]string, 0)

	for k, v := range req.Cookie.Cookies {
		if !v.Status {
			continue
		}

		value := strings.TrimSpace(v.Value)
		result := fmt.Sprintf("%s=%s", k, value)
		buffer = append(buffer, result)
	}

	if len(buffer) == 0 {
		return make([]string, 0)
	}

	return []string{
		fmt.Sprintf(`-H "Cookie: %s"`, strings.Join(buffer, "; ")),
	}
}

func headersToCurl(req *action.Request) []string {
	buffer := make([]string, 0)

	for k, h := range req.Header.Headers {
		for _, v := range h {
			if !v.Status {
				continue
			}

			value := strings.TrimSpace(v.Value)
			result := fmt.Sprintf(`-H "%s: %s"`, k, value)
			buffer = append(buffer, result)
		}
	}

	return buffer
}

func bodyToCurl(req *action.Request) []string {
	if !req.Body.Status || req.Body.ContentType == body.None {
		return make([]string, 0)
	}

	if req.Body.ContentType == body.Form {
		return formDataTocurl(req.Body)
	}

	return rawTocurl(req.Body)
}

func rawTocurl(b body.BodyRequest) []string {
	buffer := make([]string, 0)

	parameters, ok := b.Parameters[body_strategy.DOCUMENT_PARAM]
	if !ok {
		return buffer
	}

	payload, ok := parameters[body_strategy.PAYLOAD_PARAM]
	if !ok {
		return buffer
	}

	if len(payload) == 0 || payload[0].IsFile {
		return buffer
	}

	buffer = append(buffer, fmt.Sprintf("-d '%s'", payload[0].Value))

	return buffer
}

func formDataTocurl(b body.BodyRequest) []string {
	buffer := make([]string, 0)

	for k, p := range b.Parameters[body_strategy.FORM_DATA_PARAM] {
		for _, v := range p {
			if !v.Status {
				continue
			}

			if !v.IsFile {
				buffer = append(buffer, fmt.Sprintf(`-F "%s=%s"`, k, v.Value))
			} else {
				file := fmt.Sprintf(`-F "%s=@<(echo '%s' | base64 --decode);filename=%s"`, k, v.Value, v.FileName)
				buffer = append(buffer, file)
			}
		}
	}

	return buffer
}
