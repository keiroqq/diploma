package rss

type categoryAlias struct {
	categorySlug string
	rawTags      []string
}

var categoryAliases = []categoryAlias{
	{categorySlug: "backend", rawTags: []string{"backend", "бэкенд", "go", "golang", "java", "python", "api", "rest api", "микросервисы"}},
	{categorySlug: "frontend", rawTags: []string{"frontend", "фронтенд", "javascript", "typescript", "react", "vue", "ui"}},
	{categorySlug: "mobile", rawTags: []string{"mobile", "android", "ios", "flutter", "react native"}},
	{categorySlug: "gamedev", rawTags: []string{"gamedev", "game development", "игры", "unity", "unreal engine"}},
	{categorySlug: "devops", rawTags: []string{"devops", "docker", "kubernetes", "ci/cd", "linux", "администрирование"}},
	{categorySlug: "databases", rawTags: []string{"databases", "базы данных", "postgresql", "postgres", "sql", "mysql", "redis"}},
	{categorySlug: "security", rawTags: []string{"security", "infosec", "information security", "информационная безопасность", "иб", "pentest", "пентест", "криптография"}},
	{categorySlug: "ai", rawTags: []string{"ai", "artificial intelligence", "ии", "искусственный интеллект", "ml", "machine learning", "машинное обучение", "llm", "нейросети", "rag"}},
	{categorySlug: "design", rawTags: []string{"design", "дизайн", "ux", "ui/ux"}},
	{categorySlug: "management", rawTags: []string{"management", "менеджмент", "product management", "продакт-менеджмент"}},
	{categorySlug: "marketing", rawTags: []string{"marketing", "маркетинг", "growth"}},
	{categorySlug: "science", rawTags: []string{"science", "научпоп", "космос", "physics"}},
	{categorySlug: "hardware", rawTags: []string{"hardware", "железо", "электроника", "embedded", "diy"}},
}

var categorySlugByTagSlug = buildCategorySlugIndex()

func buildCategorySlugIndex() map[string]string {
	result := map[string]string{}
	for _, aliasGroup := range categoryAliases {
		for _, rawTag := range aliasGroup.rawTags {
			result[normalizeTagSlug(rawTag)] = aliasGroup.categorySlug
		}
	}
	return result
}

func categorySlugsForTags(tags []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)
	for _, tag := range tags {
		categorySlug, ok := categorySlugByTagSlug[normalizeTagSlug(tag)]
		if !ok {
			continue
		}
		if _, exists := seen[categorySlug]; exists {
			continue
		}
		seen[categorySlug] = struct{}{}
		result = append(result, categorySlug)
	}
	return result
}
