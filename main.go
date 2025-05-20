package main

import (
	"fmt"
	"log/slog"
	"os"
)

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
