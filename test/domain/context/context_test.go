package test_context

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	body_strategy "github.com/Rafael24595/go-api-core/src/domain/action/body/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-collections/collection"
)

func TestIdenfyVariables(t *testing.T) {
	source := "Lorem ipsum ${var_1} ${global.var_2} amet, ${query.var_3} adipiscing ${var_4}. Morbi eleifend odio quis ${global.var_1} commodo sodales."

	variables := context.NewContext("anonymous").IdentifyVariables("global", source)
	if len(variables) != 4 {
		t.Errorf("Found variables %d but %d expected", len(variables), 4)
	}

	variables = collection.VectorFromList(variables).Sort(func(i, j collection.Pair[string, string]) bool {
		return strings.Compare(
			fmt.Sprintf("%s.%s", i.Key(), i.Value()),
			fmt.Sprintf("%s.%s", j.Key(), j.Value())) == -1
	}).Collect()

	var0 := variables[0]
	if var0.Key() != "global" && var0.Value() != "var_1" {
		t.Errorf("Found variable %s - %s but %s - %s expected", var0.Key(), var0.Value(), "global", "var_1")
	}

	var1 := variables[1]
	if var1.Key() != "global" && var1.Value() != "var_2" {
		t.Errorf("Found variable %s - %s but %s - %s expected", var1.Key(), var1.Value(), "global", "var_2")
	}

	var2 := variables[2]
	if var2.Key() != "global" && var2.Value() != "var_4" {
		t.Errorf("Found variable %s - %s but %s - %s expected", var2.Key(), var2.Value(), "global", "var_4")
	}

	var3 := variables[3]
	if var3.Key() != "query" && var3.Value() != "var_3" {
		t.Errorf("Found variable %s - %s but %s - %s expected", var3.Key(), var3.Value(), "query", "var_3")
	}
}

func TestContextApply(t *testing.T) {
	source := "Lorem ${var_1} dolor ${global.var_2} amet, ${query.var_3} adipiscing ${var_4}. Morbi eleifend odio quis ${global.var_1} commodo ${query.var_5}."
	ctx := context.NewContext("anonymous").
		PutAll("global", map[string]context.ItemContext{
			"var_1": context.NewItemContext(0, false, true, "ipsum"),
			"var_2": context.NewItemContext(1, false, true, "sit"),
			"var_4": context.NewItemContext(2, false, true, "elit"),
		}).
		PutAll("query", map[string]context.ItemContext{
			"var_3": context.NewItemContext(0, false, true, "consectetur"),
			"var_5": context.NewItemContext(1, false, false, "sodales"),
		})

	fixSource := ctx.Apply("global", source)
	expected := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi eleifend odio quis ipsum commodo ."
	if fixSource != expected {
		t.Errorf("Found source %s but %s expected", fixSource, expected)
	}
}

func TestProcessRequest(t *testing.T) {
	var dtoRequestRaw dto.DtoRequest
	var requestExpected dto.DtoRequest

	err := json.Unmarshal(readJSON("sources/request001_raw.json"), &dtoRequestRaw)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(readJSON("sources/request001_expected.json"), &requestExpected)
	if err != nil {
		log.Panic(err)
	}

	ctx := context.NewContext("anonymous").
		PutAll("global", map[string]context.ItemContext{
			"user": context.NewItemContext(0, false, true, "Rafael24595"),
		}).
		PutAll("uri", map[string]context.ItemContext{
			"repository": context.NewItemContext(0, false, true, "go-api-core"),
		}).
		PutAll("query", map[string]context.ItemContext{
			"branch": context.NewItemContext(0, false, true, "dev"),
		}).
		PutAll("header", map[string]context.ItemContext{
			"type": context.NewItemContext(0, false, true, "application/json"),
		}).
		PutAll("payload", map[string]context.ItemContext{
			"status": context.NewItemContext(0, false, true, "true"),
		}).
		PutAll("auth", map[string]context.ItemContext{
			"pass": context.NewItemContext(0, false, true, "secret-key"),
		})

	request := context.ProcessRequest(dto.ToRequest(&dtoRequestRaw), ctx)

	found := request.Uri
	expected := requestExpected.Uri
	if found != expected {
		t.Errorf("Found source %s but %s expected", found, expected)
	}

	found = request.Query.Queries["branch"][0].Value
	expected = requestExpected.Query.Queries["branch"][0].Value
	if found != expected {
		t.Errorf("Found source %s but %s expected", found, expected)
	}

	found = request.Header.Headers["content-type"][0].Value
	expected = requestExpected.Header.Headers["content-type"][0].Value
	if found != expected {
		t.Errorf("Found source %s but %s expected", found, expected)
	}

	foundType := request.Body.ContentType
	expectedType := requestExpected.Body.ContentType
	if foundType != expectedType {
		t.Errorf("Found source %s but %s expected", foundType, expectedType)
	}

	found = string(request.Body.Parameters[body_strategy.DOCUMENT_PARAM][body_strategy.PAYLOAD_PARAM][0].Value)
	expected = string(requestExpected.Body.Parameters[body_strategy.DOCUMENT_PARAM][body_strategy.PAYLOAD_PARAM][0].Value)
	if found != expected {
		t.Errorf("Found source %s but %s expected", found, expected)
	}

	found = request.Auth.Auths["basic"].Parameters["username"]
	expected = requestExpected.Auth.Auths["basic"].Parameters["username"]
	if found != expected {
		t.Errorf("Found source %s but %s expected", found, expected)
	}

	found = request.Auth.Auths["basic"].Parameters["password"]
	expected = requestExpected.Auth.Auths["basic"].Parameters["password"]
	if found != expected {
		t.Errorf("Found source %s but %s expected", found, expected)
	}
}

func readJSON(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Panic(err)
	}

	return bytes
}
