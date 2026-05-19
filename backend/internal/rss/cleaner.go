package rss

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

type Cleaner struct {
	policy *bluemonday.Policy
}

func NewCleaner() *Cleaner {
	return &Cleaner{policy: bluemonday.StrictPolicy()}
}

func (c *Cleaner) CleanText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(value))
	if err == nil {
		value = doc.Text()
	}

	value = c.policy.Sanitize(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = strings.ReplaceAll(value, "Читать далее", "")
	value = strings.Join(strings.Fields(value), " ")
	return strings.TrimSpace(value)
}

func (c *Cleaner) ExtractImage(value string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(value))
	if err != nil {
		return ""
	}

	imageURL := ""
	doc.Find("img").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		src, ok := selection.Attr("src")
		if !ok || strings.TrimSpace(src) == "" {
			return true
		}
		imageURL = strings.TrimSpace(src)
		return false
	})

	if imageURL == "" {
		return ""
	}
	if parsed, err := url.Parse(imageURL); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return imageURL
	}
	return ""
}
