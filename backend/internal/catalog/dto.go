package catalog

import (
	"time"

	"github.com/google/uuid"
)

type DiscoverRequest struct {
	PageURL string `json:"page_url"`
}

type DiscoverResponse struct {
	PageURL string `json:"page_url"`
	FeedURL string `json:"feed_url"`
	Title   string `json:"title,omitempty"`
}

type ConnectCatalogSourcesRequest struct {
	SourceIDs []string `json:"source_ids"`
}

type ConnectedCatalogSource struct {
	CatalogSourceID string    `json:"catalog_source_id"`
	SourceID        uuid.UUID `json:"source_id"`
	FeedSourceID    uuid.UUID `json:"feed_source_id"`
	Title           string    `json:"title"`
	PageURL         string    `json:"page_url"`
	FeedURL         string    `json:"feed_url"`
	CreatedAt       time.Time `json:"created_at"`
}

type ConnectCatalogSourcesResponse struct {
	Connected []ConnectedCatalogSource `json:"connected"`
}
