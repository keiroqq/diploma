package catalog

import (
	"net/url"
	"strings"
	"testing"
)

func TestDiscoverFindsRSSLink(t *testing.T) {
	pageURL, err := url.Parse("https://habr.com/ru/flows/backend/news/")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	resp, err := discoverFromHTML(
		pageURL,
		strings.NewReader(`<html><head><title>Backend news</title><link rel="alternate" type="application/rss+xml" href="/rss/backend/"></head></html>`),
	)
	if err != nil {
		t.Fatalf("discoverFromHTML returned error: %v", err)
	}

	if resp.FeedURL != "https://habr.com/rss/backend/" {
		t.Fatalf("FeedURL = %q", resp.FeedURL)
	}
	if resp.Title != "Backend news" {
		t.Fatalf("Title = %q", resp.Title)
	}
}

func TestDiscoverReturnsErrorWithoutRSSLink(t *testing.T) {
	pageURL, err := url.Parse("https://habr.com/ru/flows/backend/news/")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	_, err = discoverFromHTML(pageURL, strings.NewReader(`<html><head><title>No feed</title></head></html>`))
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrRSSNotFound {
		t.Fatalf("err = %v, want %v", err, ErrRSSNotFound)
	}
}
