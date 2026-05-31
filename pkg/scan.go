package pkg

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type MatchResult struct {
	Name        string
	Description string
	Confidence  string
	Match       string
}

type NamedRegexp struct {
	Name        string
	Description string
	Confidence  string
	Re          *regexp.Regexp
}

type Pattern struct {
	Regex       string `yaml:"regex"`
	Description string `yaml:"description"`
	Confidence  string `yaml:"confidence"`
}

type RegexConfig struct {
	Patterns map[string]Pattern `yaml:"patterns"`
}

func LoadRegexFile(path string) (map[string]Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var config RegexConfig

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.Patterns, nil
}

func CompileRegexes(patterns map[string]Pattern) []NamedRegexp {
	out := make([]NamedRegexp, 0)

	for name, pattern := range patterns {
		r, err := regexp.Compile(pattern.Regex)
		if err != nil {
			// skip invalid regex
			continue
		}

		out = append(out, NamedRegexp{
			Name:        name,
			Description: pattern.Description,
			Confidence:  pattern.Confidence,
			Re:          r,
		})
	}

	return out
}

func ProcessURL(client *http.Client, url string, regs []NamedRegexp) ([]MatchResult, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	text := string(body)

	var results []MatchResult

	for _, line := range strings.Split(text, "\n") {
		for _, reg := range regs {
			if m := reg.Re.FindString(line); m != "" {
				results = append(results, MatchResult{
					Name:        reg.Name,
					Description: reg.Description,
					Confidence:  reg.Confidence,
					Match:       m,
				})
			}
		}
	}

	return results, nil
}
