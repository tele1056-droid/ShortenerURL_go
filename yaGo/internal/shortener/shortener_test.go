package shortener_test

import (
	"github.com/tele1056-droid/yaGo/internal/shortener"
	"testing"
)

func TestShorten(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		longURL string
		want    string
	}{
		// TODO: Add test cases.
		{"Request 1", "URL1", "1"},
		{"Request 2", "URL2", "2"},
		{"Request 3", "URL3", "3"},
		{"Request 4", "URL4", "4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortener.Shorten()
			// TODO: update the condition below to compare got with tt.want.

			if got != tt.want {
				t.Errorf("Shorten() = %v, want %v", got, tt.want)
			}
		})
	}
}
