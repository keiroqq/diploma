package items

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
	htmlnode "golang.org/x/net/html"

	"github.com/keiro/content-digest/backend/internal/models"
)

var (
	errArticleContentUnavailable = errors.New("article content unavailable")
	vedomostiBodyExpr            = regexp.MustCompile(`"body":"((?:\\.|[^"])*)"`)
	habrMetricExpr               = regexp.MustCompile(`(?i)^\+?\d+(?:[.,]\d+)?\s*(?:мин|k|к|m|м)?$`)
	habrTimeExpr                 = regexp.MustCompile(`\b\d{1,2}:\d{2}\b`)
)

type ArticleReader struct {
	client *http.Client
	policy *bluemonday.Policy
}

func NewArticleReader(client *http.Client) *ArticleReader {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("h2", "h3", "p", "br", "ul", "ol", "li", "blockquote", "pre", "code", "strong", "b", "em", "i")
	policy.AllowAttrs("href").OnElements("a")
	policy.AllowStandardURLs()
	policy.RequireNoFollowOnLinks(true)
	policy.RequireNoReferrerOnLinks(true)
	policy.AddTargetBlankToFullyQualifiedLinks(true)

	return &ArticleReader{
		client: client,
		policy: policy,
	}
}

func (r *ArticleReader) Fetch(ctx context.Context, item models.FeedItem) (string, error) {
	if r == nil || r.client == nil {
		return "", errArticleContentUnavailable
	}
	if strings.TrimSpace(item.URL) == "" {
		return "", errArticleContentUnavailable
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", "ContentDigest/1.0 (+https://localhost)")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%w: status %d", errArticleContentUnavailable, resp.StatusCode)
	}

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	raw := string(rawBytes)
	if strings.TrimSpace(raw) == "" {
		return "", errArticleContentUnavailable
	}

	host := articleHost(item)
	htmlContent, err := r.extract(host, raw)
	if err != nil {
		return "", err
	}
	return htmlContent, nil
}

func (r *ArticleReader) Sanitize(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return normalizeArticleHTML(r.policy.Sanitize(value))
}

func (r *ArticleReader) SanitizeForItem(item models.FeedItem) string {
	value := strings.TrimSpace(item.ContentHTML)
	if value == "" {
		return ""
	}
	if strings.Contains(articleHost(item), "habr.com") {
		value = cleanHabrArticleHTML(value)
	}
	return r.Sanitize(value)
}

func (r *ArticleReader) FallbackHTML(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	paragraphs := strings.Split(text, "\n")
	var builder strings.Builder
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		builder.WriteString("<p>")
		builder.WriteString(html.EscapeString(paragraph))
		builder.WriteString("</p>")
	}
	return builder.String()
}

func (r *ArticleReader) ShouldFetch(item models.FeedItem) bool {
	if strings.TrimSpace(item.ContentHTML) == "" {
		return true
	}

	host := articleHost(item)
	content := strings.ToLower(item.ContentHTML)
	if strings.Contains(host, "habr.com") {
		return strings.Contains(content, "хабы:") ||
			strings.Contains(content, "теги:") ||
			strings.Contains(content, "время прочтения") ||
			strings.Contains(content, "tm-article-presenter")
	}
	return false
}

func (r *ArticleReader) extract(host string, raw string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(raw))
	if err != nil {
		return "", err
	}

	var extracted string
	switch {
	case strings.Contains(host, "habr.com"):
		extracted = r.htmlFromHabr(doc)
	case strings.Contains(host, "sports.ru"):
		extracted = r.htmlFromSports(doc)
	case strings.Contains(host, "vedomosti.ru"):
		extracted = r.htmlFromVedomosti(raw)
	case strings.Contains(host, "kommersant.ru"):
		extracted = r.htmlFromFirst(doc, ".article_text, .doc__text, article")
	}
	if extracted == "" {
		extracted = r.htmlFromFirst(doc, "[itemprop='articleBody'], article")
	}
	if extracted == "" {
		return "", errArticleContentUnavailable
	}
	return extracted, nil
}

func (r *ArticleReader) htmlFromHabr(doc *goquery.Document) string {
	for _, selector := range []string{
		"#post-content-body .article-formatted-body",
		"#post-content-body",
		".tm-article-body .article-formatted-body",
		".tm-article-presenter__body .article-formatted-body",
		".tm-article-presenter__body",
		".tm-article-body",
		"[data-test-id='article-content']",
		"[itemprop='articleBody']",
		".article-formatted-body",
		"[data-gallery-root] .article-formatted-body",
		"article",
	} {
		selection := doc.Find(selector).First()
		if selection.Length() == 0 {
			continue
		}
		selection = selection.Clone()
		cleanHabrSelection(selection)
		if htmlContent := r.htmlFromSelection(selection); htmlContent != "" {
			return htmlContent
		}
	}
	return ""
}

func (r *ArticleReader) htmlFromFirst(doc *goquery.Document, selector string) string {
	selection := doc.Find(selector).First()
	if selection.Length() == 0 {
		return ""
	}
	return r.htmlFromSelection(selection.Clone())
}

func (r *ArticleReader) htmlFromSports(doc *goquery.Document) string {
	var builder strings.Builder
	doc.Find(".sb-paragraph").Each(func(_ int, selection *goquery.Selection) {
		text := strings.Join(strings.Fields(selection.Text()), " ")
		if text == "" {
			return
		}
		body, err := selection.Html()
		if err != nil || strings.TrimSpace(body) == "" {
			body = html.EscapeString(text)
		}
		builder.WriteString("<p>")
		builder.WriteString(body)
		builder.WriteString("</p>")
	})
	return r.Sanitize(builder.String())
}

func (r *ArticleReader) htmlFromVedomosti(raw string) string {
	matches := vedomostiBodyExpr.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		body, err := strconv.Unquote(`"` + match[1] + `"`)
		if err != nil {
			continue
		}
		if strings.TrimSpace(body) == "" {
			continue
		}
		builder.WriteString("<p>")
		builder.WriteString(body)
		builder.WriteString("</p>")
	}
	return r.Sanitize(builder.String())
}

func (r *ArticleReader) htmlFromSelection(selection *goquery.Selection) string {
	selection.Find("script, style, noscript, iframe, svg, form, button, nav, aside, footer, header, .comments, .ad, .adv, .banner").Remove()
	body, err := selection.Html()
	if err != nil || strings.TrimSpace(body) == "" {
		body = selection.Text()
	}
	return r.Sanitize(body)
}

func cleanHabrArticleHTML(value string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(value))
	if err != nil {
		return value
	}

	selection := firstHabrContentSelection(doc)
	if selection.Length() == 0 {
		selection = doc.Find("body").First()
	}
	if selection.Length() == 0 {
		return value
	}

	selection = selection.Clone()
	cleanHabrSelection(selection)

	body, err := selection.Html()
	if err != nil || strings.TrimSpace(body) == "" {
		body = selection.Text()
	}
	return body
}

func firstHabrContentSelection(doc *goquery.Document) *goquery.Selection {
	for _, selector := range []string{
		"#post-content-body .article-formatted-body",
		"#post-content-body",
		".tm-article-body .article-formatted-body",
		".tm-article-presenter__body .article-formatted-body",
		".tm-article-presenter__body",
		".tm-article-body",
		"[data-test-id='article-content']",
		"[itemprop='articleBody']",
		".article-formatted-body",
	} {
		selection := doc.Find(selector).First()
		if selection.Length() > 0 {
			return selection
		}
	}
	return doc.Find("article").First()
}

func cleanHabrSelection(selection *goquery.Selection) {
	selection.Find(strings.Join([]string{
		"h1",
		".tm-article-presenter__header",
		".tm-article-presenter__meta",
		".tm-article-presenter__meta-list",
		".tm-article-presenter__footer",
		".tm-article-presenter__labels",
		".tm-article-reading-time",
		".tm-article-complexity",
		".tm-article-snippet__hubs",
		".tm-article-snippet__labels",
		".tm-article-snippet__stats",
		".tm-article-snippet__meta",
		".tm-article-snippet__title",
		".tm-article-snippet__footer",
		".tm-separated-list",
		".tm-tags-list",
		".tm-hubs-list",
		".tm-user-info",
		".tm-votes-meter",
		".tm-data-icons",
		".tm-article-sticky-panel",
		"[data-test-id='article-author']",
		"[data-test-id='article-meta']",
		"[class*='article__meta']",
		"[class*='article__footer']",
		"[class*='article__hubs']",
		"[class*='article__tags']",
		"[class*='article__stats']",
	}, ", ")).Remove()

	selection.Find("p, div, span").Each(func(_ int, node *goquery.Selection) {
		if isHabrChromeText(compactText(node.Text())) {
			node.Remove()
		}
	})
	removeHabrTailSections(selection)
}

func removeHabrTailSections(selection *goquery.Selection) {
	for _, node := range selection.Nodes {
		removeHabrTailFromNode(node)
	}
}

func removeHabrTailFromNode(parent *htmlnode.Node) bool {
	if parent == nil {
		return false
	}

	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		if isHabrTailMarker(compactText(nodeText(child))) {
			removeNodeAndFollowing(parent, child)
			return true
		}

		if removeHabrTailFromNode(child) {
			if child.Parent == parent {
				removeNodeAndFollowing(parent, child.NextSibling)
			}
			return true
		}
	}
	return false
}

func removeNodeAndFollowing(parent *htmlnode.Node, start *htmlnode.Node) {
	for node := start; node != nil; {
		next := node.NextSibling
		parent.RemoveChild(node)
		node = next
	}
}

func nodeText(node *htmlnode.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == htmlnode.TextNode {
		return node.Data
	}

	var builder strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		builder.WriteString(nodeText(child))
		builder.WriteString(" ")
	}
	return builder.String()
}

func isHabrChromeText(text string) bool {
	lower := strings.ToLower(text)
	runeCount := len([]rune(text))

	if text == "" {
		return false
	}
	if runeCount <= 24 && habrMetricExpr.MatchString(text) {
		return true
	}
	if runeCount <= 90 && habrTimeExpr.MatchString(text) &&
		(strings.Contains(lower, "сегодня") || strings.Contains(lower, "вчера") || strings.Contains(lower, "назад") || strings.Contains(lower, " в ")) {
		return true
	}
	if runeCount <= 180 && strings.Contains(lower, "блог компании") {
		return true
	}
	return false
}

func isHabrTailMarker(text string) bool {
	lower := strings.Trim(strings.ToLower(text), ": ")
	return lower == "теги" ||
		lower == "хабы" ||
		strings.HasPrefix(lower, "теги:") ||
		strings.HasPrefix(lower, "хабы:")
}

func compactText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func normalizeArticleHTML(value string) string {
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(value))
	if err != nil {
		return strings.TrimSpace(value)
	}
	if strings.TrimSpace(doc.Text()) == "" {
		return ""
	}

	htmlValue, err := doc.Find("body").Html()
	if err != nil {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(htmlValue)
}

func articleHost(item models.FeedItem) string {
	candidates := []string{item.URL, item.Source.URL, item.Source.FeedURL}
	for _, candidate := range candidates {
		parsed, err := url.Parse(strings.TrimSpace(candidate))
		if err == nil && parsed.Hostname() != "" {
			return strings.ToLower(parsed.Hostname())
		}
	}
	return ""
}
