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
	//makeing sure were crawling the same domain

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

	// //normalize the rawCurrentURL

	// normCurr, err := normalizeURL(rawCurrentURL)
	// if err != nil {
	// 	slog.Error("Normalizing error", "error", err)
	// 	return
	// }
	// fmt.Printf("---- normalized version of the current url: %s\n ----", normCurr)
	// //Check if we've already crawled the rawCurrentURL

	if _, exists := (*pages)[rawCurrentURL]; exists {
		(*pages)[rawCurrentURL]++
		return
	} else {
		(*pages)[rawCurrentURL] = 1
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
		fmt.Printf("extracted link: %s \n", v)
	}
	*pool = append(*pool, currPageURLs...)
	*pool = (*pool)[1:]
	if len(*pool) == 0 {
		return
	}
	newCurrUrl := (*pool)[0]
	crawlPage(rawBaseURL, newCurrUrl, pages, pool)
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

func realtiveToAbs(relativePath, rootUrl string) (absUrl string) {
	root := strings.TrimSuffix(rootUrl, "/")
	absUrl = root + relativePath
	return
}

func normalizeURL(rawUrl string) (result string, err error) {
	actual, err := url.Parse(rawUrl)
	if err != nil {
		log.Fatal(err)
	}
	host := actual.Host
	path := strings.TrimSuffix(actual.Path, "/")
	result = host + path
	return result, nil
}

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	urls := []string{}
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		log.Fatal(err)
		return []string{}, err
	}
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.A {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if strings.HasPrefix(a.Val, "/") {
						url := realtiveToAbs(a.Val, rawBaseURL)
						urls = append(urls, url)
						break
					} else {
						urls = append(urls, a.Val)
						break
					}
				}
			}
		}
	}
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
