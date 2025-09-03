package estimate

import "testing"

func TestHasEstimate(t *testing.T) {
	tests := []struct {
		body string
		want bool
	}{
		{"", false},
		{"No estimate here", false},
		{"Estimate: 2 days", true},
		{"estimate: 1 day", true},
		{"Estimate: 2.5 days", true},
		{"Estimate 2 days", false},
		{"Estimate: two days", false},
		{"ETA: 2d", false},
	}
	for _, tt := range tests {
		if got := HasEstimate(tt.body); got != tt.want {
			t.Errorf("HasEstimate(%q) = %v; want %v", tt.body, got, tt.want)
		}
	}
}
