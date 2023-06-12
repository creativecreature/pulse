package filetypes_test

import (
	"testing"

	"code-harvest.conner.dev/pkg/filetypes"
)

func TestGet(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
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
		got, _ := filetypes.Type(test.filename)
		if got != test.expected {
			t.Errorf("Get(%s) = %s; wanted %s", test.filename, got, test.expected)
		}
	}
}
