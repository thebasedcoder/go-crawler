package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

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

func main() {
	urls := []string{
		"https://example.com/path/",
		"example.com/no/scheme",
		"http://www.test.com/",
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
