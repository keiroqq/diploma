package rss

import "testing"

func TestCategorySlugsForTagsNormalizesAliases(t *testing.T) {
	cases := []struct {
		name string
		tags []string
		want string
	}{
		{name: "russian ai", tags: []string{"искусственный интеллект"}, want: "ai"},
		{name: "short ai", tags: []string{"ИИ"}, want: "ai"},
		{name: "machine learning", tags: []string{"machine learning"}, want: "ai"},
		{name: "backend", tags: []string{"Go"}, want: "backend"},
		{name: "security", tags: []string{"информационная безопасность"}, want: "security"},
		{name: "sports", tags: []string{"спорт"}, want: "sports"},
		{name: "football", tags: []string{"Футбол"}, want: "football"},
		{name: "combat sports", tags: []string{"UFC"}, want: "combat-sports"},
		{name: "figure skating", tags: []string{"фигурное катание"}, want: "figure-skating"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := categorySlugsForTags(tc.tags)
			if len(got) != 1 {
				t.Fatalf("len(got) = %d, want 1: %#v", len(got), got)
			}
			if got[0] != tc.want {
				t.Fatalf("got[0] = %q, want %q", got[0], tc.want)
			}
		})
	}
}
