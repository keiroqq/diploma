package catalog

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var ErrRSSNotFound = errors.New("rss feed not found")
var ErrPageFetchFailed = errors.New("page fetch failed")

type Discoverer struct {
	client *http.Client
}

func NewDiscoverer(client *http.Client) *Discoverer {
	if client == nil {
		client = http.DefaultClient
	}
	return &Discoverer{client: client}
}

func (d *Discoverer) Discover(ctx context.Context, pageURL string) (DiscoverResponse, error) {
	parsed, err := validatePageURL(pageURL)
	if err != nil {
		return DiscoverResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return DiscoverResponse{}, err
	}
	req.Header.Set("User-Agent", "content-digest-app/0.1")

	resp, err := d.client.Do(req)
	if err != nil {
		return DiscoverResponse{}, errors.Join(ErrPageFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return DiscoverResponse{}, ErrPageFetchFailed
	}

	return discoverFromHTML(parsed, resp.Body)
}

func discoverFromHTML(pageURL *url.URL, body io.Reader) (DiscoverResponse, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return DiscoverResponse{}, err
	}

	feedURL := ""
	title := strings.TrimSpace(doc.Find("title").First().Text())
	doc.Find(`link[type="application/rss+xml"]`).EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		href, ok := selection.Attr("href")
		if !ok || strings.TrimSpace(href) == "" {
			return true
		}
		resolved, err := resolveURL(pageURL, href)
		if err != nil {
			return true
		}
		feedURL = resolved
		return false
	})

	if feedURL == "" {
		return DiscoverResponse{}, ErrRSSNotFound
	}

	return DiscoverResponse{
		PageURL: pageURL.String(),
		FeedURL: feedURL,
		Title:   title,
	}, nil
}

func validatePageURL(value string) (*url.URL, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("page_url is required")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("page_url must use http or https")
	}
	if parsed.Host == "" {
		return nil, errors.New("page_url host is required")
	}
	return parsed, nil
}

func resolveURL(base *url.URL, value string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", err
	}
	return base.ResolveReference(parsed).String(), nil
}
