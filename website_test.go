package enola

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestWebsiteUnmarshalErrorMsgString(t *testing.T) {
	raw := `{"errorType":"message","errorMsg":"Page Not Found","url":"https://example.com/{}"}`
	var w Website
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	want := []string{"Page Not Found"}
	if !reflect.DeepEqual(w.ErrorMessages, want) {
		t.Fatalf("got %v, want %v", w.ErrorMessages, want)
	}
}

func TestWebsiteUnmarshalErrorMsgArray(t *testing.T) {
	raw := `{"errorType":"message","errorMsg":["a","b"],"url":"https://example.com/{}"}`
	var w Website
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	want := []string{"a", "b"}
	if !reflect.DeepEqual(w.ErrorMessages, want) {
		t.Fatalf("got %v, want %v", w.ErrorMessages, want)
	}
}

func TestWebsiteUnmarshalErrorCodeInt(t *testing.T) {
	raw := `{"errorType":"status_code","errorCode":404,"url":"https://example.com/{}"}`
	var w Website
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	want := []int{404}
	if !reflect.DeepEqual(w.ErrorCodes, want) {
		t.Fatalf("got %v, want %v", w.ErrorCodes, want)
	}
}

func TestWebsiteUnmarshalErrorCodeArray(t *testing.T) {
	raw := `{"errorType":"status_code","errorCode":[404,500],"url":"https://example.com/{}"}`
	var w Website
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	want := []int{404, 500}
	if !reflect.DeepEqual(w.ErrorCodes, want) {
		t.Fatalf("got %v, want %v", w.ErrorCodes, want)
	}
}
