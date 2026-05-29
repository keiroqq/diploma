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

func TestCatalogCategorySlugsForSourcesIncludesVedomostiTopics(t *testing.T) {
	got := catalogCategorySlugsForSources([]feedSourceRef{
		{
			Name:    "Ведомости: Политика",
			URL:     "https://www.vedomosti.ru/rss/rubric/politics.xml",
			FeedURL: "https://www.vedomosti.ru/rss/rubric/politics.xml",
		},
		{
			Name:    "Ведомости: Технологии",
			URL:     "https://www.vedomosti.ru/rss/rubric/technology.xml",
			FeedURL: "https://www.vedomosti.ru/rss/rubric/technology.xml",
		},
	})

	want := map[string]bool{
		"politics":   true,
		"technology": true,
	}
	for _, slug := range got {
		delete(want, slug)
	}
	if len(want) > 0 {
		t.Fatalf("missing vedomosti category slugs: %#v; got %#v", want, got)
	}
}

func TestCatalogCategorySlugsForSourcesIncludesKommersantTopics(t *testing.T) {
	got := catalogCategorySlugsForSources([]feedSourceRef{
		{
			Name:    "Коммерсантъ: В мире",
			URL:     "https://www.kommersant.ru/RSS/section-world.xml",
			FeedURL: "https://www.kommersant.ru/RSS/section-world.xml",
		},
		{
			Name:    "Коммерсантъ: Происшествия",
			URL:     "https://www.kommersant.ru/RSS/section-accidents.xml",
			FeedURL: "https://www.kommersant.ru/RSS/section-accidents.xml",
		},
		{
			Name:    "Коммерсантъ: Самара",
			URL:     "https://www.kommersant.ru/rss/regions/samara_all.xml",
			FeedURL: "https://www.kommersant.ru/rss/regions/samara_all.xml",
		},
	})

	want := map[string]bool{
		"world":     true,
		"accidents": true,
		"regions":   true,
	}
	for _, slug := range got {
		delete(want, slug)
	}
	if len(want) > 0 {
		t.Fatalf("missing kommersant category slugs: %#v; got %#v", want, got)
	}
}
