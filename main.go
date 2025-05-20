package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		slog.Error("Usage: go run main.go <url>")
		os.Exit(1)
	}

	startURL := args[0]
	rawURL, err := url.Parse(startURL)
	if err != nil {
		panic("Invalid start URL")
	}
	slog.Info("Starting crawl", "url", rawURL)
	fmt.Println("-------------------------")

	urls := []string{startURL}
	pages := make(map[string]int)

	crawlPage(startURL, startURL, &pages, &urls)

	fmt.Println("\n--- Crawl Report ---")
	for url, count := range pages {
		fmt.Printf("Seen %d time(s): %s\n", count, url)
	}
	slog.Info("Finished crawl", "pages", len(pages))
}
