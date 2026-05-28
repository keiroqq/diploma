package categories

import "testing"

func TestCatalogCategorySlugsForSourcesIncludesSportsTopics(t *testing.T) {
	got := catalogCategorySlugsForSources([]feedSourceRef{
		{
			Name:    "Sports.ru: Футбол",
			URL:     "https://www.sports.ru/rss/rubric/208.xml",
			FeedURL: "https://www.sports.ru/rss/rubric/208.xml",
		},
		{
			Name:    "Sports.ru: Бокс/MMA/UFC",
			URL:     "https://www.sports.ru/rss/rubric/213.xml",
			FeedURL: "https://www.sports.ru/rss/rubric/213.xml",
		},
	})

	want := map[string]bool{
		"sports":        true,
		"football":      true,
		"combat-sports": true,
	}
	for _, slug := range got {
		delete(want, slug)
	}
	if len(want) > 0 {
		t.Fatalf("missing sports category slugs: %#v; got %#v", want, got)
	}
}
