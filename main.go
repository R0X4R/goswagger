package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"gopkg.in/yaml.v3"

	optionspkg "github.com/R0X4R/goswagger/pkg"
)

//go:embed regex.yaml
var defaultRegexYAML []byte

func ensureDefaultRegexFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	// if file doesn't exist > write directly
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, defaultRegexYAML, 0o644)
	}

	// load existing user file
	userPatterns, err := optionspkg.LoadRegexFile(path)
	if err != nil {
		return err
	}

	var embedded struct {
		Patterns map[string]optionspkg.Pattern `yaml:"patterns"`
	}

	if err := yaml.Unmarshal(defaultRegexYAML, &embedded); err != nil {
		return err
	}

	defaultPatterns := embedded.Patterns

	// merge only missing patterns
	changed := false
	for name, pattern := range defaultPatterns {
		if _, exists := userPatterns[name]; !exists {
			userPatterns[name] = pattern
			changed = true
		}
	}

	if !changed {
		return nil
	}

	// rebuild yaml structure
	out := struct {
		Patterns map[string]optionspkg.Pattern `yaml:"patterns"`
	}{
		Patterns: userPatterns,
	}

	data, err := yaml.Marshal(out)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func main() {
	defaultRegexFile, err := optionspkg.DefaultRegexFilePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to resolve config directory:", err)
		os.Exit(1)
	}
	options, err := optionspkg.ParseOptions()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if options.RegexFile == defaultRegexFile {
		if err := ensureDefaultRegexFile(options.RegexFile); err != nil {
			fmt.Fprintln(os.Stderr, "failed to seed default regex file:", err)
			os.Exit(1)
		}
	}

	if options.BaseURL != "" {
		optionspkg.SwaggerBaseURLFormat = options.BaseURL
	}
	// set global maxPages
	if options.MaxPages >= 0 {
		optionspkg.MaxFetchPages = options.MaxPages
	}

	regexMap, err := optionspkg.LoadRegexFile(options.RegexFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load regex file:", err)
		os.Exit(1)
	}

	compiled := optionspkg.CompileRegexes(regexMap)

	urls, err := optionspkg.GetURLs(options.Query)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to fetch urls:", err)
		os.Exit(1)
	}

	if len(urls) == 0 {
		fmt.Println("No results")
		return
	}

	sem := make(chan struct{}, options.Threads)
	var wg sync.WaitGroup

	client := &http.Client{Timeout: 15 * time.Second}

	var out *os.File
	var outMu sync.Mutex
	if options.OutputFile != "" {
		if err := os.MkdirAll(filepath.Dir(options.OutputFile), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "failed to create output directory:", err)
			os.Exit(1)
		}

		f, err := os.OpenFile(
			options.OutputFile,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to open output file:", err)
			os.Exit(1)
		}
		out = f
		defer out.Close()
	}

	for _, u := range urls {
		wg.Add(1)
		sem <- struct{}{}
		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }()
			matches, err := optionspkg.ProcessURL(client, url, compiled)
			if err != nil {
				return
			}
			if len(matches) > 0 {
				for _, m := range matches {
					// Minimal output: url | pattern | match
					// format pattern name: convert underscores to spaces and uppercase
					prettyName := strings.ToUpper(strings.ReplaceAll(m.Name, "_", " "))

					line := fmt.Sprintf(
						"[%s] [%s] %s [%s]\n",
						strings.ToUpper(m.Confidence),
						prettyName,
						url,
						m.Match,
					)
					fmt.Print(line)
					if out != nil {
						outMu.Lock()
						// out.WriteString(line)
						if _, err := out.WriteString(line); err != nil {
							fmt.Fprintln(os.Stderr, "failed to write output:", err)
						}
						outMu.Unlock()
					}
				}
			}
		}(u)
	}

	wg.Wait()
}
