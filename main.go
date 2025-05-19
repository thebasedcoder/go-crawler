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

var inputBody = `
<html>
        <body>
                <a href="/path/one">
                        <span>Boot.dev</span>
                </a>
                <a href="https://other.com/path/one">
                        <span>Boot.dev</span>
                </a>
        </body>
</html>
`

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

	body, err := getHTML(rawUrl)
	if err != nil {
		os.Exit(1)
	}
	fmt.Println(body)
	fmt.Println("-------------------------")
	slog.Info("Starting the crawl", "Target", rawUrl)
	fmt.Println("-------------------------")
	urls, err := getURLsFromHTML(inputBody, rawUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, u := range urls {
		norm, err := normalizeURL(u)
		if err != nil {
			fmt.Printf("Error normalizing %s: %v\n", u, err)
			continue
		}
		fmt.Printf("%-25s => %s\n", u, norm)
	}

}
