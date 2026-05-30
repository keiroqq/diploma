package items

import (
	"strings"
	"testing"

	"github.com/keiro/content-digest/backend/internal/models"
)

func TestArticleReaderExtractsProviderContent(t *testing.T) {
	reader := NewArticleReader(nil)

	cases := []struct {
		name string
		host string
		html string
		want string
	}{
		{
			name: "habr",
			host: "habr.com",
			html: `<html><body><article><div class="tm-article-presenter__header"><h1>Лишний заголовок</h1><span>2 мин</span><span>8.1K</span><div>Блог компании Selectel</div></div><div id="post-content-body"><div class="article-formatted-body"><p>Первый абзац Хабра.</p><p>Второй абзац.</p></div></div><footer>Хабы:<ul><li>Backend</li></ul></footer></article></body></html>`,
			want: "Первый абзац Хабра.",
		},
		{
			name: "sports",
			host: "www.sports.ru",
			html: `<html><body><p class="sb-paragraph">Первый абзац Sports.</p><p class="sb-paragraph">Второй абзац.</p></body></html>`,
			want: "Первый абзац Sports.",
		},
		{
			name: "vedomosti",
			host: "www.vedomosti.ru",
			html: `<html><body><script>{"body":"Первый абзац Ведомостей.","body":"Второй абзац."}</script></body></html>`,
			want: "Первый абзац Ведомостей.",
		},
		{
			name: "kommersant",
			host: "www.kommersant.ru",
			html: `<html><body><div class="article_text"><p>Первый абзац Коммерсанта.</p><p>Второй абзац.</p></div></body></html>`,
			want: "Первый абзац Коммерсанта.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := reader.extract(tc.host, tc.html)
			if err != nil {
				t.Fatalf("extract() error = %v", err)
			}
			if !strings.Contains(got, tc.want) {
				t.Fatalf("extract() = %q, want text %q", got, tc.want)
			}
			if strings.Contains(got, "Хабы:") || strings.Contains(got, "Лишний заголовок") {
				t.Fatalf("extract() contains non-reader chrome: %q", got)
			}
			if strings.Contains(got, "Блог компании") || strings.Contains(got, "8.1K") {
				t.Fatalf("extract() contains habr metadata: %q", got)
			}
		})
	}
}

func TestArticleReaderSanitizesStoredHabrChrome(t *testing.T) {
	reader := NewArticleReader(nil)
	item := models.FeedItem{
		URL: "https://habr.com/ru/articles/123/",
		ContentHTML: `<article>
			<div><img src="https://example.test/avatar.png" alt=""> promo_speech вчера в 14:05</div>
			<h1>Дублирующийся заголовок</h1>
			<p>3 мин</p>
			<p>8.1K</p>
			<p>Блог компании Selectel Компьютерное железо IT-инфраструктура</p>
			<p>Первый полезный абзац.</p>
			<p>Второй полезный абзац.</p>
			Теги:
			<ul><li>selectel</li><li>хранение данных</li></ul>
			<p>Хабы:</p>
			<ul><li>Блог компании Selectel</li></ul>
			<p>+15</p>
			<p>4</p>
		</article>`,
	}

	got := reader.SanitizeForItem(item)
	for _, unwanted := range []string{
		"Дублирующийся заголовок",
		"promo_speech",
		"3 мин",
		"8.1K",
		"Блог компании",
		"Теги:",
		"Хабы:",
		"+15",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("SanitizeForItem() contains %q in %q", unwanted, got)
		}
	}
	if !strings.Contains(got, "Первый полезный абзац.") || !strings.Contains(got, "Второй полезный абзац.") {
		t.Fatalf("SanitizeForItem() = %q, want useful article paragraphs", got)
	}
}
