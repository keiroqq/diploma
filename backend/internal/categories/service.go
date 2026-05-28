package categories

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/keiro/content-digest/backend/internal/catalog"
	httpx "github.com/keiro/content-digest/backend/internal/http"
	"github.com/keiro/content-digest/backend/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]CategoryResponse, error) {
	records, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]CategoryResponse, 0, len(records))
	for _, record := range records {
		resp = append(resp, categoryResponse(record))
	}
	return resp, nil
}

func (s *Service) ListByFeed(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) ([]CategoryResponse, error) {
	exists, err := s.repo.FeedExistsForUser(ctx, feedID, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, httpx.ErrNotFound
	}

	sources, err := s.repo.ListFeedSources(ctx, feedID, userID)
	if err != nil {
		return nil, err
	}

	records, err := s.repo.ListBySlugs(ctx, catalogCategorySlugsForSources(sources))
	if err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Name < records[j].Name
	})

	resp := make([]CategoryResponse, 0, len(records))
	for _, record := range records {
		resp = append(resp, categoryResponse(record))
	}
	return resp, nil
}

func catalogCategorySlugsForSources(sources []feedSourceRef) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)
	for _, topic := range catalog.Topics() {
		for _, source := range topic.Sources {
			if !matchesCatalogSource(sources, source) {
				continue
			}

			for _, tag := range source.Tags {
				slug, ok := categorySlugBySourceTag[normalizeSourceTag(tag)]
				if !ok {
					continue
				}
				if _, exists := seen[slug]; exists {
					continue
				}

				seen[slug] = struct{}{}
				result = append(result, slug)
			}
		}
	}
	return result
}

func matchesCatalogSource(sources []feedSourceRef, catalogSource catalog.CatalogSource) bool {
	catalogPageURL := normalizeURL(catalogSource.PageURL)
	catalogTitle := normalizeSourceTag(catalogSource.Title)

	for _, source := range sources {
		if normalizeURL(source.URL) == catalogPageURL || normalizeURL(source.FeedURL) == catalogPageURL {
			return true
		}

		if catalogTitle != "" && strings.Contains(normalizeSourceTag(source.Name), catalogTitle) {
			return true
		}
	}
	return false
}

func normalizeURL(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func normalizeSourceTag(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

var categorySlugBySourceTag = map[string]string{
	"admin":             "devops",
	"администрирование": "devops",
	"ai":                "ai",
	"analytics":         "databases",
	"android":           "mobile",
	"api":               "backend",
	"backend":           "backend",
	"bi":                "databases",
	"ci/cd":             "devops",
	"chips":             "hardware",
	"data":              "databases",
	"databases":         "databases",
	"design":            "design",
	"devops":            "devops",
	"diy":               "hardware",
	"docker":            "devops",
	"electronics":       "hardware",
	"embedded":          "hardware",
	"flutter":           "mobile",
	"frontend":          "frontend",
	"gamedev":           "gamedev",
	"games":             "gamedev",
	"go":                "backend",
	"growth":            "marketing",
	"hardware":          "hardware",
	"infosec":           "security",
	"ios":               "mobile",
	"java":              "backend",
	"javascript":        "frontend",
	"kubernetes":        "devops",
	"linux":             "devops",
	"llm":               "ai",
	"management":        "management",
	"marketing":         "marketing",
	"ml":                "ai",
	"mobile":            "mobile",
	"pc":                "hardware",
	"pentest":           "security",
	"physics":           "science",
	"postgresql":        "databases",
	"product":           "management",
	"python":            "backend",
	"react":             "frontend",
	"sales":             "marketing",
	"science":           "science",
	"security":          "security",
	"servers":           "devops",
	"space":             "science",
	"sql":               "databases",
	"team":              "management",
	"typescript":        "frontend",
	"ui":                "design",
	"unity":             "gamedev",
	"unreal":            "gamedev",
	"ux":                "design",
	"нейросети":         "ai",
}

func categoryResponse(category models.Category) CategoryResponse {
	return CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		CreatedAt:   category.CreatedAt,
	}
}
