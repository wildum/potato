package main

import "testing"

func TestNormalizeRouteName(t *testing.T) {
	tests := []struct {
		name   string
		route  string
		method string
		want   string
	}{
		{
			name:   "removes duplicated method",
			route:  "GET /potatoes",
			method: "GET",
			want:   "/potatoes",
		},
		{
			name:   "keeps route without method prefix",
			route:  "/inventory",
			method: "GET",
			want:   "/inventory",
		},
		{
			name:   "handles extra whitespace",
			route:  "GET    /analytics",
			method: "GET",
			want:   "/analytics",
		},
		{
			name:   "case insensitive comparison",
			route:  "get /health",
			method: "GET",
			want:   "/health",
		},
		{
			name:   "different method leaves route untouched",
			route:  "POST /recipes",
			method: "GET",
			want:   "POST /recipes",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeRouteName(tt.route, tt.method); got != tt.want {
				t.Fatalf("normalizeRouteName(%q, %q) = %q, want %q", tt.route, tt.method, got, tt.want)
			}
		})
	}
}
