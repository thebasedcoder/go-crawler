package main

import (
	"fmt"
	"log/slog"
	"net/url"
)

func crawlPage(rawBaseURL, rawCurrentURL string, pages *map[string]int, pool *[]string) {
	base, err := url.Parse(rawBaseURL)
	if err != nil {
		slog.Error("Invalid base URL", "error", err)
		return
	}
	curr, err := url.Parse(rawCurrentURL)
	if err != nil {
		slog.Error("Invalid current URL", "error", err)
		return
	}

	if curr.Host != base.Host {
		return
	}

	fmt.Printf("Crawling the url: %s\n", rawCurrentURL)

	normCurr, err := normalizeURL(rawCurrentURL)
	if err != nil {
		slog.Error("Normalization failed", "url", rawCurrentURL, "error", err)
		return
	}

	if _, exists := (*pages)[normCurr]; exists {
		(*pages)[normCurr]++
		return
	}
	(*pages)[normCurr] = 1

	htmlBody, err := getHTML(rawCurrentURL)
	if err != nil {
		slog.Error("Failed to get HTML", "url", rawCurrentURL, "error", err)
	}
	links, err := getURLsFromHTML(htmlBody, rawBaseURL)
	if err != nil {
		slog.Error("HTML parsing failed", "url", rawCurrentURL, "error", err)
		return
	}

	for _, link := range links {
		fmt.Printf("found: %s\n", link)
	}
	for _, link := range links {
		norm, err := normalizeURL(link)
		if err != nil {
			continue
		}
		parsed, err := url.Parse(norm)
		if err != nil || parsed.Host != base.Host {
			continue
		}
		if _, seen := (*pages)[norm]; !seen {
			*pool = append(*pool, norm)
		}
	}

	*pool = (*pool)[1:]

	if len(*pool) == 0 {
		return
	}

	nextURL := (*pool)[0]

	crawlPage(rawBaseURL, nextURL, pages, pool)
}
