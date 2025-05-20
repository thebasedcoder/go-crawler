package main

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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
