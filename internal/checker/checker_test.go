package checker

import (
	"net/http"
	"testing"
)

func TestStatusCodeDetector(t *testing.T) {
	d := statusCodeDetector{}
	tests := []struct {
		name   string
		resp   Response
		target Target
		want   bool
	}{
		{
			name:   "200 ok",
			resp:   Response{StatusCode: http.StatusOK},
			target: Target{},
			want:   true,
		},
		{
			name:   "404 not found",
			resp:   Response{StatusCode: http.StatusNotFound},
			target: Target{},
			want:   false,
		},
		{
			name:   "error code match",
			resp:   Response{StatusCode: 404},
			target: Target{ErrorCodes: []int{404}},
			want:   false,
		},
		{
			name:   "200 with error codes excluding 404",
			resp:   Response{StatusCode: http.StatusOK},
			target: Target{ErrorCodes: []int{404}},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.Detect(tt.resp, tt.target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageDetector(t *testing.T) {
	d := messageDetector{}
	tests := []struct {
		name   string
		resp   Response
		target Target
		want   bool
	}{
		{
			name:   "single message absent",
			resp:   Response{Body: []byte("welcome profile")},
			target: Target{ErrorMessages: []string{"not found"}},
			want:   true,
		},
		{
			name:   "single message present",
			resp:   Response{Body: []byte("user not found")},
			target: Target{ErrorMessages: []string{"not found"}},
			want:   false,
		},
		{
			name:   "array any match",
			resp:   Response{Body: []byte("<title>404 Not Found</title>")},
			target: Target{ErrorMessages: []string{"something went wrong", "404 Not Found"}},
			want:   false,
		},
		{
			name:   "array none match",
			resp:   Response{Body: []byte("profile page")},
			target: Target{ErrorMessages: []string{"something went wrong", "404 Not Found"}},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.Detect(tt.resp, tt.target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponseURLDetector(t *testing.T) {
	d := responseURLDetector{}
	tests := []struct {
		name   string
		resp   Response
		target Target
		want   bool
	}{
		{
			name:   "redirected to error url",
			resp:   Response{FinalURL: "https://example.com/error404.aspx"},
			target: Target{ErrorURL: "https://example.com/error404.aspx"},
			want:   false,
		},
		{
			name:   "profile url kept",
			resp:   Response{FinalURL: "https://example.com/user/alice"},
			target: Target{ErrorURL: "https://example.com/error404.aspx"},
			want:   true,
		},
		{
			name:   "empty error url falls back to status",
			resp:   Response{StatusCode: http.StatusOK, FinalURL: "https://example.com/user/alice"},
			target: Target{},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.Detect(tt.resp, tt.target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistryLookup(t *testing.T) {
	types := []string{"status_code", "message", "response_url"}
	for _, errorType := range types {
		t.Run(errorType, func(t *testing.T) {
			d, ok := Lookup(errorType)
			if !ok {
				t.Fatalf("detector not registered for %q", errorType)
			}
			if d == nil {
				t.Fatal("detector is nil")
			}
		})
	}

	if _, ok := Lookup("unknown"); ok {
		t.Fatal("expected unknown error type to be missing")
	}
}
