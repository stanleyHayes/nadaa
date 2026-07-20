package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONCapsBodySize(t *testing.T) {
	oversized := `{"note":"` + strings.Repeat("a", 1<<20) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(oversized))
	recorder := httptest.NewRecorder()
	var target struct {
		Note string `json:"note"`
	}
	if err := DecodeJSON(recorder, request, &target); err == nil {
		t.Fatal("expected bodies over 1 MiB to fail decoding")
	}
}

func TestDecodeJSONAcceptsSmallBody(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"note":"ok"}`))
	recorder := httptest.NewRecorder()
	var target struct {
		Note string `json:"note"`
	}
	if err := DecodeJSON(recorder, request, &target); err != nil || target.Note != "ok" {
		t.Fatalf("expected a small body to decode, got error=%v target=%#v", err, target)
	}
}
