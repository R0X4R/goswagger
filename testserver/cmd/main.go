package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func specBodyForPath(path string) string {
	switch path {
	case "/specs/1":
		return "This is a fake spec containing API key AIza12345678901234567890123456789012345 and heroku id e601d11c-48d4-4c0e-93e7-b6fde873b820."
	case "/specs/2":
		return "basic YWRtaW46YWRtaW4= with s3://sample-bucket and access_token$production$12345678$0123456789abcdef0123456789abcdef"
	case "/specs/3":
		return "Bearer eyJhbGciOiAiSFMyNTYifQ.eyJzdWIiOiAiZGV2In0.signature and AC1234567890ABCDEF1234567890ABCDEF"
	case "/specs/edge":
		return strings.Join([]string{
			"-----BEGIN RSA PRIVATE KEY-----",
			"MIIEowIBAAKCAQEAtest",
			"-----END RSA PRIVATE KEY-----",
			"md5 5f4dcc3b5aa765d61d8327deb882cf99",
		}, "\n")
	default:
		return "This is a fake spec containing API key AIza12345678901234567890123456789012345 for testing."
	}
}

func main() {
	mux := http.NewServeMux()
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("TESTSERVER_MODE")))
	if mode == "" {
		mode = "multi"
	}
	port := os.Getenv("TESTSERVER_PORT")
	if port == "" {
		port = "8081"
	}
	basePath := os.Getenv("TESTSERVER_BASE_PATH")
	if basePath == "" {
		basePath = "/apiproxy/specs"
	}

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "0"
		}

		pathsByPage := map[string][]string{
			"0": {"/specs/1", "/specs/2"},
			"1": {"/specs/3"},
			"2": {"/specs/edge"},
		}
		if mode == "basic" {
			pathsByPage = map[string][]string{"0": {"/specs/1"}}
		}

		paths, ok := pathsByPage[page]
		if !ok {
			paths = []string{}
		}

		apis := make([]map[string]any, 0, len(paths))
		for _, specPath := range paths {
			apis = append(apis, map[string]any{
				"properties": []any{map[string]any{"url": fmt.Sprintf("http://%s%s", r.Host, specPath)}},
			})
		}

		totalCount := 300
		if mode == "basic" {
			totalCount = 100
		}
		if override := os.Getenv("TESTSERVER_TOTAL_COUNT"); override != "" {
			if value, err := strconv.Atoi(override); err == nil {
				totalCount = value
			}
		}

		writeJSON(w, map[string]any{
			"totalCount": totalCount,
			"apis":       apis,
		})
	})

	for _, path := range []string{"/specs/1", "/specs/2", "/specs/3", "/specs/edge"} {
		path := path
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if delay := os.Getenv("TESTSERVER_DELAY_MS"); delay != "" {
				if ms, err := strconv.Atoi(delay); err == nil && ms > 0 {
					time.Sleep(time.Duration(ms) * time.Millisecond)
				}
			}
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte(specBodyForPath(path)))
		})
	}

	addr := ":" + port
	_ = http.ListenAndServe(addr, mux)
}
