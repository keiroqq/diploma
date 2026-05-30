package catalog

import "testing"

func TestHabrSourceUsesDirectRSSFeedURL(t *testing.T) {
	cases := []struct {
		name string
		page string
		want string
	}{
		{
			name: "flow",
			page: "https://habr.com/ru/flows/frontend/news/",
			want: "https://habr.com/ru/rss/flows/frontend/news/?fl=ru",
		},
		{
			name: "hub",
			page: "https://habr.com/ru/hubs/devops/news/",
			want: "https://habr.com/ru/rss/hubs/devops/news/?fl=ru",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := habrSource("test", "Test", "Test source", tc.page, nil)
			if source.FeedURL != tc.want {
				t.Fatalf("FeedURL = %q, want %q", source.FeedURL, tc.want)
			}
		})
	}
}
