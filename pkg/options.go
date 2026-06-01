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
	NoColor    bool
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
		flagSet.StringVarP(&options.Query, "query", "q", "", "\tSearch query (required)"),
		flagSet.StringVarP(&options.RegexFile, "regex-file", "r", defaultRegexFile, "\tPath to regex YAML file"),
	)

	flagSet.CreateGroup("runtime", "Runtime",
		flagSet.IntVarP(&options.Threads, "threads", "t", 25, "\tNumber of worker threads"),
		flagSet.StringVarP(&options.BaseURL, "base-url", "b", "", "\tOverride SwaggerHub base URL format (use %s and %d for query and page)"),
		flagSet.IntVarP(&options.MaxPages, "max-pages", "m", 1, "\tMaximum number of pages to fetch (0 = all)"),
	)

	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&options.OutputFile, "output", "o", "", "\tWrite matches to this file (txt). Appends if file exists"),
		flagSet.BoolVarP(&options.NoColor, "no-color", "n", false, "\tDisable colored terminal output"),
	)

	if err := flagSet.Parse(); err != nil {
		return nil, err
	}

	if options.Query == "" {
		return nil, fmt.Errorf("[!] GOSWAGGER - Query is required (use -q)")
	}

	return options, nil
}
