package git_test

import (
	"testing"

	"github.com/creativecreature/pulse/git"
)

type testCase struct {
	filename string
	expected string
}

func TestGetFileType(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{"styles.css", "css"},
		{"index.js", "javascript"},
		{"component.tsx", "typescript"},
		{"init.lua", "lua"},
		{"Dockerfile", "Docker"},
		{"docker-compose.yaml", "Docker"},
		{"docker-compose.yml", "Docker"},
		{"Makefile", "Makefile"},
	}

	for _, test := range tests {
		got, _ := git.Filetype(test.filename)
		if got != test.expected {
			t.Errorf("Get(%s) = %s; wanted %s", test.filename, got, test.expected)
		}
	}
}
