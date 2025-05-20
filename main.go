package main

import (
	"fmt"
	"io"
	"log"
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
		log.Fatal(err)
		return
	}
	curr, err := url.Parse(rawCurrentURL)
	if err != nil {
		log.Fatal(err)
		return
	}

	if curr.Host != base.Host {
		return
	}
	fmt.Printf("Crawling the url: %s\n", rawCurrentURL)
	normCurr, err := normalizeURL(rawCurrentURL)

	if _, exists := (*pages)[normCurr]; exists {
		(*pages)[normCurr]++
		return
	} else {
		(*pages)[normCurr] = 1
	}

	currBody, err := getHTML(rawCurrentURL)
	if err != nil {
		slog.Error("Error while getting the page", "error", err, "url", rawCurrentURL)
	}
	currPageURLs, err := getURLsFromHTML(currBody, rawBaseURL)
	if err != nil {
		slog.Error("Error while parsing HTML", "URL", rawCurrentURL)

		return
	}
	for _, v := range currPageURLs {
		fmt.Printf("founde: %s\n", v)
	}
	for _, v := range currPageURLs {
		norm, err := normalizeURL(v)
		if err == nil {
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
		slog.Error("Operation failed", "error", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 400 {
		slog.Error("status code error", "code", resp.StatusCode, "url", resp.Request.URL)
		return "", fmt.Errorf("failed with status code of: %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		slog.Error("Unexpected content type", "content-type", contentType)
		return "", fmt.Errorf("Unexpected content-type: %s, expexted text/html", contentType)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Operation failed", "error", err)
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

func normalizeURL(rawUrl string) (string, error) {
	parsed, err := url.Parse(rawUrl)
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
		log.Fatal(err)
		return []string{}, err
	}
	var urls []string
	extractLinks(doc, rawBaseURL, &urls)
	return urls, nil
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		slog.Warn("no url provided", "u", "NONE")
		os.Exit(1)
	}
	if len(args) > 1 {
		slog.Warn("TOo many args", "required argument", "u")
		os.Exit(1)
	}
	rawUrl := os.Args[1]

	fmt.Println("-------------------------")
	slog.Info("Starting the crawl", "Target", rawUrl)
	fmt.Println("-------------------------")

	urls := []string{rawUrl}
	rawCurrentUrl := urls[0]
	pages := make(map[string]int)
	crawlPage(rawUrl, rawCurrentUrl, &pages, &urls)
	for key, value := range pages {
		fmt.Printf("the %s seen %d\n", key, value)
	}
	slog.Info("finnished execution", "status", "success")
}
