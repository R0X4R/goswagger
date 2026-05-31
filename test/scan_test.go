package pkg_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	pkg "github.com/R0X4R/goswagger/pkg"
	"gopkg.in/yaml.v3"
)

func TestLoadAndCompileRegexes(t *testing.T) {
	m, err := pkg.LoadRegexFile("../regex.yaml")
	if err != nil {
		t.Fatalf("failed to load regex.yaml: %v", err)
	}
	if len(m) == 0 {
		t.Fatalf("expected regex map to be non-empty")
	}
	regs := pkg.CompileRegexes(m)
	if len(regs) == 0 {
		t.Fatalf("expected compiled regex map to be non-empty")
	}
}

func TestLoadRegexFileAllowsListValues(t *testing.T) {
	data := []byte(`amazon_aws_url:
  - 's3\\.amazonaws\\.com[/]+'
  - '[a-zA-Z0-9_-]*\\.s3\\.amazonaws\\.com'
heroku_api_key: '[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}'
`)
	raw := make(map[string]any)
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal inline yaml: %v", err)
	}

	patterns := make(map[string][]string)
	for name, value := range raw {
		switch typed := value.(type) {
		case string:
			patterns[name] = []string{typed}
		case []any:
			for _, item := range typed {
				if pattern, ok := item.(string); ok {
					patterns[name] = append(patterns[name], pattern)
				}
			}
		}
	}

	if got := len(patterns["amazon_aws_url"]); got != 2 {
		t.Fatalf("expected 2 patterns for amazon_aws_url, got %d", got)
	}
	if got := len(patterns["heroku_api_key"]); got != 1 {
		t.Fatalf("expected 1 pattern for heroku_api_key, got %d", got)
	}

	compiled := pkg.CompileRegexes(patterns)
	if len(compiled) != 3 {
		t.Fatalf("expected 3 compiled regexes, got %d", len(compiled))
	}
}

func TestProcessURLMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("some text AIza12345678901234567890123456789012345 other text"))
	}))
	defer srv.Close()

	regs := []pkg.NamedRegexp{
		{Name: "google_api", Re: regexp.MustCompile(`AIza[0-9A-Za-z-_]{35}`)},
	}

	client := &http.Client{}
	matches, err := pkg.ProcessURL(client, srv.URL, regs)
	if err != nil {
		t.Fatalf("processURL error: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("expected at least one match, got none")
	}
	if matches[0].Name != "google_api" {
		t.Fatalf("expected google_api match, got %s", matches[0].Name)
	}
}