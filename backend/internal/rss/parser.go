package rss

import (
	"context"
	"net/http"

	"github.com/mmcdole/gofeed"
)

type Parser struct {
	parser *gofeed.Parser
}

func NewParser(client *http.Client) *Parser {
	parser := gofeed.NewParser()
	if client != nil {
		parser.Client = client
	}
	return &Parser{parser: parser}
}

func (p *Parser) ParseURL(ctx context.Context, feedURL string) (*gofeed.Feed, error) {
	return p.parser.ParseURLWithContext(feedURL, ctx)
}
