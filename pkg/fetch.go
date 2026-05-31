package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SwaggerBaseURLFormat is the format string for the SwaggerHub search endpoint.
// Tests may override this variable to point to a local httptest server.
var SwaggerBaseURLFormat = "https://app.swaggerhub.com/apiproxy/specs?sort=BEST_MATCH&order=DESC&query=%s&page=%d&limit=100"

// MaxFetchPages limits how many pages the scanner will fetch (0 = all pages)
var MaxFetchPages = 1

// GetURLs fetches SwaggerHub search results and returns discovered URLs
func GetURLs(query string) ([]string, error) {
	base := SwaggerBaseURLFormat
	client := &http.Client{Timeout: 15 * time.Second}

	// page 0
	reqURL := fmt.Sprintf(base, query, 0)
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var apiResp struct {
		TotalCount int `json:"totalCount"`
		Apis       []struct {
			Properties []struct {
				URL string `json:"url"`
			} `json:"properties"`
		} `json:"apis"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	urls := extractURLs(apiResp)
	pages := apiResp.TotalCount / 100
	if MaxFetchPages > 0 && pages > MaxFetchPages {
		pages = MaxFetchPages
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for p := 1; p <= pages; p++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			req := fmt.Sprintf(base, query, page)
			r, err := client.Get(req)
			if err != nil {
				return
			}
			defer r.Body.Close()
			var pr struct {
				Apis []struct {
					Properties []struct {
						URL string `json:"url"`
					} `json:"properties"`
				} `json:"apis"`
			}
			if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
				return
			}
			mu.Lock()
			urls = append(urls, extractURLs(pr)...) // reuse helper
			mu.Unlock()
		}(p)
	}
	wg.Wait()

	// filter empty
	var out []string
	for _, u := range urls {
		if strings.TrimSpace(u) != "" {
			out = append(out, u)
		}
	}
	return out, nil
}

// extractURLs extracts urls from a decoded response shape
func extractURLs(v interface{}) []string {
	var urls []string
	data, _ := json.Marshal(v)
	var tmp struct {
		Apis []struct {
			Properties []struct {
				URL string `json:"url"`
			} `json:"properties"`
		} `json:"apis"`
	}
	_ = json.Unmarshal(data, &tmp)
	for _, api := range tmp.Apis {
		for _, p := range api.Properties {
			urls = append(urls, p.URL)
		}
	}
	return urls
}
