package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
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
