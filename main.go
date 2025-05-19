package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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
