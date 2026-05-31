package pkg_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pkg "github.com/R0X4R/goswagger/pkg"
)

func TestGetURLsSinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "0" {
			resp := map[string]interface{}{
				"totalCount": 1,
				"apis": []interface{}{map[string]interface{}{
					"properties": []interface{}{map[string]interface{}{"url": "http://example.com/spec1"}},
				}},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]interface{}{"apis": []interface{}{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	pkg.SwaggerBaseURLFormat = srv.URL + "?query=%s&page=%d&limit=100"

	urls, err := pkg.GetURLs("testquery")
	if err != nil {
		t.Fatalf("getURLs error: %v", err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected 1 url, got %d", len(urls))
	}
	if urls[0] != "http://example.com/spec1" {
		t.Fatalf("unexpected url: %s", urls[0])
	}
}