package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/projectdiscovery/goflags"
)

type Options struct {
	Query      string
	RegexFile  string
	Threads    int
	BaseURL    string
	MaxPages   int
	OutputFile string
}

func DefaultRegexFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "goswagger", "regex.yaml"), nil
}

func ParseOptions() (*Options, error) {
	os.Args[0] = "goswagger"

	defaultRegexFile, err := DefaultRegexFilePath()
	if err != nil {
		return nil, err
	}

	options := &Options{}
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`GOSWAGGER - A minimal SwaggerHub OSINT scanner focusing on concise output and configurable regex patterns. Inspired by SwaggerSpy`)

	flagSet.CreateGroup("input", "Input",
		flagSet.StringVarP(&options.Query, "query", "q", "", "Search query (required)"),
		flagSet.StringVarP(&options.RegexFile, "regex-file", "r", defaultRegexFile, "Path to regex YAML file"),
	)

	flagSet.CreateGroup("runtime", "Runtime",
		flagSet.IntVarP(&options.Threads, "threads", "t", 25, "Number of worker threads"),
		flagSet.StringVarP(&options.BaseURL, "base-url", "b", "", "Override SwaggerHub base URL format (use %s and %d for query and page)"),
		flagSet.IntVarP(&options.MaxPages, "max-pages", "m", 1, "Maximum number of pages to fetch (0 = all)"),
	)

	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&options.OutputFile, "output", "o", "", "Write matches to this file (txt). Appends if file exists"),
	)

	if err := flagSet.Parse(); err != nil {
		return nil, err
	}

	if options.Query == "" {
		return nil, fmt.Errorf("query is required (use -q)")
	}

	return options, nil
}
