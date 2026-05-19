package rss

import (
	"context"

	"github.com/mmcdole/gofeed"
)

type Parser struct {
	parser *gofeed.Parser
}

func NewParser() *Parser {
	return &Parser{parser: gofeed.NewParser()}
}

func (p *Parser) ParseURL(ctx context.Context, feedURL string) (*gofeed.Feed, error) {
	return p.parser.ParseURLWithContext(feedURL, ctx)
}
