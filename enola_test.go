package enola

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func collectResults(ch <-chan Result) []Result {
	var out []Result
	for r := range ch {
		out = append(out, r)
	}
	return out
}

func TestCheckStatusCodeFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	data := map[string]Website{
		"TestSite": {
			ErrorType: "status_code",
			URL:       server.URL + "/{}",
		},
	}

	e, err := New(WithData(data), WithHTTPClient(server.Client()), WithConcurrency(1))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ch, err := e.Check(context.Background(), "alice")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	results := collectResults(ch)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if !results[0].Found {
		t.Fatalf("expected found result, got %+v", results[0])
	}
}

func TestCheckMessageNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("user not found"))
	}))
	defer server.Close()

	data := map[string]Website{
		"TestSite": {
			ErrorType:     "message",
			ErrorMessages: []string{"not found"},
			URL:           server.URL + "/{}",
		},
	}

	e, err := New(WithData(data), WithHTTPClient(server.Client()), WithConcurrency(1))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ch, err := e.Check(context.Background(), "alice")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	results := collectResults(ch)
	if len(results) != 1 || results[0].Found {
		t.Fatalf("expected not found result, got %+v", results)
	}
}

func TestCheckRegexInvalid(t *testing.T) {
	data := map[string]Website{
		"TestSite": {
			ErrorType:  "status_code",
			URL:        "https://example.com/{}",
			RegexCheck: "^[a-z]+$",
		},
	}

	e, err := New(WithData(data), WithConcurrency(1))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ch, err := e.Check(context.Background(), "Alice123")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	results := collectResults(ch)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Status != StatusInvalid {
		t.Fatalf("expected invalid status, got %+v", results[0])
	}
}

func TestCheckSiteFilterNotFound(t *testing.T) {
	e, err := New(WithData(map[string]Website{
		"Twitter": {ErrorType: "status_code", URL: "https://example.com/{}"},
	}))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = e.SetSite("does-not-exist").Check(context.Background(), "alice")
	if !errors.Is(err, ErrSiteNotFound) {
		t.Fatalf("expected ErrSiteNotFound, got %v", err)
	}
}

func TestCheckChannelCloses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	data := map[string]Website{
		"A": {ErrorType: "status_code", URL: server.URL + "/a/{}"},
		"B": {ErrorType: "status_code", URL: server.URL + "/b/{}"},
	}

	e, err := New(WithData(data), WithHTTPClient(server.Client()), WithConcurrency(2))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ch, err := e.Check(context.Background(), "alice")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	results := collectResults(ch)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestCheckRequestPayloadAndHeaders(t *testing.T) {
	var gotMethod, gotContentType string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		gotBody = body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	data := map[string]Website{
		"TestSite": {
			ErrorType:      "status_code",
			URL:            server.URL + "/{}",
			URLProbe:       server.URL + "/probe",
			RequestMethod:  http.MethodPost,
			RequestPayload: map[string]any{"username": "{}"},
			Headers:        map[string]string{"Content-Type": "application/json"},
		},
	}

	e, err := New(WithData(data), WithHTTPClient(server.Client()), WithConcurrency(1))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ch, err := e.Check(context.Background(), "alice")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	_ = collectResults(ch)

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content-type = %q, want application/json", gotContentType)
	}
	if string(gotBody) != `{"username":"alice"}` {
		t.Fatalf("body = %q", gotBody)
	}
}
