package test_utils

import (
	"reflect"
	"testing"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

type CaseSplitCommand struct {
	name string
	data string
	want []string
	fail bool
}

func TestSplitCommand(t *testing.T) {
	tests := []CaseSplitCommand{
		{
			name: "Simple command",
			data: `curl https://example.com`,
			want: []string{"curl", "https://example.com"},
		},
		{
			name: "Command with flags",
			data: `curl -X POST -H "Content-Type: application/json" https://api.example.com`,
			want: []string{"curl", "-X", "POST", "-H", "Content-Type: application/json", "https://api.example.com"},
		},
		{
			name: "Command with escaped quotes",
			data: `echo \"Hello World\"`,
			want: []string{"echo", `"Hello`, `World"`},
		},
		{
			name: "Command with single quotes",
			data: `echo 'Hello World'`,
			want: []string{"echo", "Hello World"},
		},
		{
			name: "Command with mixed quotes",
			data: `curl -d '{"name":"John"}'`,
			want: []string{"curl", "-d", `{"name":"John"}`},
		},
		{
			name: "Command with nested escaped quotes",
			data: `curl -H "Authorization: Bearer \"abc123\"" https://api.example.com`,
			want: []string{"curl", "-H", `Authorization: Bearer "abc123"`, "https://api.example.com"},
		},
		{
			name: "Command with multiple spaces",
			data: `curl    -X    GET     https://example.com`,
			want: []string{"curl", "-X", "GET", "https://example.com"},
		},
		{
			name: "Command with escaped space",
			data: `curl https://example.com/some\ path`,
			want: []string{"curl", "https://example.com/some path"},
		},
		{
			name: "Command with newline and tabs",
			data: "curl\t-X POST\nhttps://api.example.com",
			want: []string{"curl", "-X", "POST", "https://api.example.com"},
		},
		{
			name: "Command with unclosed quote",
			data: `curl -d '{"name":"John}`,
			fail: true,
		},
		{
			name: "Command ending with quoted argument",
			data: `echo "done"`,
			want: []string{"echo", "done"},
		},
		{
			name: "Command with empty quoted argument",
			data: `cmd "" ' '`,
			want: []string{"cmd", "", " "},
		},
		{
			name: "Command with empty quoted argument at end",
			data: `cmd " " ''`,
			want: []string{"cmd", " ", ""},
		},
		{
			name: "Command with escaped backslash",
			data: `cmd C:\\Users\\Admin`,
			want: []string{"cmd", `C:\Users\Admin`},
		},
	}

	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			run(t, data)
		})
	}
}

func run(t *testing.T, test CaseSplitCommand) {
	result, err := utils.SplitCommand(test.data)
	if (err != nil) != test.fail {
		t.Fatalf("expected error=%v, got err=%v", test.fail, err)
	}

	if !reflect.DeepEqual(result, test.want) && !test.fail {
		t.Errorf("Found %#v, but %#v expected", result, test.want)
	}
}
