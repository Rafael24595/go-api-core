package test_context

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain/context"
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
	source := "Lorem ${var_1} dolor ${global.var_2} amet, ${query.var_3} adipiscing ${var_4}. Morbi eleifend odio quis ${global.var_1} commodo sodales."
	context := context.NewContext("anonymous").
		PutAll("global", map[string]string{
			"var_1": "ipsum",
			"var_2": "sit",
			"var_4": "elit",
		}).
		PutAll("query", map[string]string{
			"var_3": "consectetur",
		})

	fixSource := context.Apply("global", source)
	expected := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi eleifend odio quis ipsum commodo sodales."
	if fixSource != expected {
		t.Errorf("Found source %s but %s expected", fixSource, expected)
	}
}