package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

//crawl page func

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
		fmt.Printf("founde: %s\n", link)
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

//crawl page func

func getHTML(rawURL string) (string, error) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    60 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 400 {
		return "", fmt.Errorf("status code %d", resp.StatusCode)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return "", fmt.Errorf("expected text/html, got %s", resp.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil

}

func relativeToAbs(relativePath, rootUrl string) string {
	base, err := url.Parse(rootUrl)
	if err != nil {
		return ""
	}
	ref, err := url.Parse(relativePath)
	if err != nil {
		return ""
	}
	return base.ResolveReference(ref).String()
}

func normalizeURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.Fragment = ""
	parsed.RawQuery = ""
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	return parsed.String(), nil
}

func extractLinks(n *html.Node, base string, urls *[]string) {
	if n.Type == html.ElementNode && n.DataAtom == atom.A {
		for _, a := range n.Attr {
			if a.Key == "href" {
				var abs string
				if strings.HasPrefix(a.Val, "/") || strings.HasPrefix(a.Val, "./") {
					abs = relativeToAbs(a.Val, base)
				} else if strings.HasPrefix(a.Val, "http") {
					abs = a.Val
				}
				if abs != "" {
					*urls = append(*urls, abs)
				}
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractLinks(c, base, urls)
	}
}

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}
	var urls []string
	extractLinks(doc, rawBaseURL, &urls)
	return urls, nil
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		slog.Error("Usage: go run main.go <url>")
		os.Exit(1)
	}

	rawURL := args[0]

	slog.Info("Starting crawl", "url", rawURL)
	fmt.Println("-------------------------")

	urls := []string{rawURL}
	pages := make(map[string]int)

	crawlPage(rawURL, rawURL, &pages, &urls)

	fmt.Println("\n--- Crawl Report ---")
	for url, count := range pages {
		fmt.Printf("Seen %d time(s): %s\n", count, url)
	}
	slog.Info("Finished crawl", "pages", len(pages))
}
