package catalog

type Topic struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Sources     []CatalogSource `json:"sources"`
}

type CatalogSource struct {
	ID          string   `json:"id"`
	Provider    string   `json:"provider"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	PageURL     string   `json:"page_url"`
	FeedURL     string   `json:"feed_url,omitempty"`
	Tags        []string `json:"tags"`
}

func Topics() []Topic {
	return []Topic{
		{
			ID:          "development",
			Title:       "Разработка",
			Description: "Новости по backend, frontend, мобильной разработке, геймдеву и языкам программирования.",
			Sources: []CatalogSource{
				habrSource("habr-backend-news", "Бэкенд", "Новости backend-разработки на Хабре.", "https://habr.com/ru/flows/backend/news/", []string{"backend", "api", "go", "java", "python"}),
				habrSource("habr-frontend-news", "Фронтенд", "Новости frontend-разработки на Хабре.", "https://habr.com/ru/flows/frontend/news/", []string{"frontend", "javascript", "typescript", "react"}),
				habrSource("habr-mobile-news", "Мобильная разработка", "Новости мобильной разработки на Хабре.", "https://habr.com/ru/flows/mobile_development/news/", []string{"mobile", "android", "ios", "flutter"}),
				habrSource("habr-gamedev-news", "GameDev", "Новости разработки игр на Хабре.", "https://habr.com/ru/flows/gamedev/news/", []string{"gamedev", "games", "unity", "unreal"}),
			},
		},
		{
			ID:          "infrastructure-data",
			Title:       "Инфраструктура и данные",
			Description: "Новости администрирования, DevOps, баз данных, аналитики и инфраструктуры.",
			Sources: []CatalogSource{
				habrSource("habr-admin-news", "Администрирование", "Новости системного администрирования на Хабре.", "https://habr.com/ru/flows/admin/news/", []string{"admin", "linux", "servers"}),
				habrSource("habr-devops-news", "DevOps", "Новости DevOps и эксплуатации на Хабре.", "https://habr.com/ru/hubs/devops/news/", []string{"devops", "docker", "kubernetes", "ci/cd"}),
				habrSource("habr-databases-news", "Базы данных", "Новости баз данных и хранения данных на Хабре.", "https://habr.com/ru/hubs/databases/news/", []string{"databases", "postgresql", "sql"}),
				habrSource("habr-analytics-news", "Аналитика данных", "Новости аналитики и работы с данными на Хабре.", "https://habr.com/ru/flows/analytics/news/", []string{"analytics", "data", "bi"}),
			},
		},
		{
			ID:          "security-ai",
			Title:       "Безопасность и AI",
			Description: "Новости информационной безопасности, искусственного интеллекта и машинного обучения.",
			Sources: []CatalogSource{
				habrSource("habr-security-news", "Информационная безопасность", "Новости ИБ на Хабре.", "https://habr.com/ru/flows/information_security/news/", []string{"security", "infosec", "pentest"}),
				habrSource("habr-ai-ml-news", "AI и ML", "Новости искусственного интеллекта и машинного обучения на Хабре.", "https://habr.com/ru/flows/ai_and_ml/news/", []string{"ai", "ml", "llm", "нейросети"}),
			},
		},
		{
			ID:          "product-business",
			Title:       "Продукт и бизнес",
			Description: "Новости дизайна, менеджмента, маркетинга и продуктовой разработки.",
			Sources: []CatalogSource{
				habrSource("habr-design-news", "Дизайн", "Новости дизайна интерфейсов и продуктов на Хабре.", "https://habr.com/ru/flows/design/news/", []string{"design", "ui", "ux"}),
				habrSource("habr-management-news", "Менеджмент", "Новости управления проектами и командами на Хабре.", "https://habr.com/ru/flows/management/news/", []string{"management", "product", "team"}),
				habrSource("habr-marketing-news", "Маркетинг", "Новости IT-маркетинга на Хабре.", "https://habr.com/ru/flows/marketing/news/", []string{"marketing", "growth", "sales"}),
			},
		},
		{
			ID:          "science-hardware",
			Title:       "Научпоп и железо",
			Description: "Новости научпопа, hardware, электроники и инженерных тем.",
			Sources: []CatalogSource{
				habrSource("habr-popsci-news", "Научпоп", "Научно-популярные новости на Хабре.", "https://habr.com/ru/flows/popsci/news/", []string{"science", "space", "physics"}),
				habrSource("habr-hardware-news", "Железо", "Новости hardware и компьютерного железа на Хабре.", "https://habr.com/ru/hubs/hardware/news/", []string{"hardware", "pc", "chips"}),
				habrSource("habr-diy-news", "DIY", "Новости DIY, электроники и инженерных проектов на Хабре.", "https://habr.com/ru/hubs/diy/news/", []string{"diy", "electronics", "embedded"}),
			},
		},
		{
			ID:          "sports",
			Title:       "Спорт",
			Description: "Новости футбола, хоккея, баскетбола, автоспорта, единоборств и других видов спорта.",
			Sources: []CatalogSource{
				sportsSource("sports-football-news", "Футбол", "Новости футбола на Sports.ru.", "https://www.sports.ru/rss/rubric/208.xml", []string{"football", "спорт", "футбол"}),
				sportsSource("sports-hockey-news", "Хоккей", "Новости хоккея на Sports.ru.", "https://www.sports.ru/rss/rubric/209.xml", []string{"hockey", "спорт", "хоккей"}),
				sportsSource("sports-basketball-news", "Баскетбол", "Новости баскетбола на Sports.ru.", "https://www.sports.ru/rss/rubric/210.xml", []string{"basketball", "спорт", "баскетбол"}),
				sportsSource("sports-formula-1-news", "Формула-1", "Новости Формулы-1 на Sports.ru.", "https://www.sports.ru/rss/rubric/211.xml", []string{"formula-1", "спорт", "автоспорт"}),
				sportsSource("sports-tennis-news", "Теннис", "Новости тенниса на Sports.ru.", "https://www.sports.ru/rss/rubric/212.xml", []string{"tennis", "спорт", "теннис"}),
				sportsSource("sports-fighting-news", "Бокс/MMA/UFC", "Новости бокса, MMA и UFC на Sports.ru.", "https://www.sports.ru/rss/rubric/213.xml", []string{"boxing", "mma", "ufc", "спорт"}),
				sportsSource("sports-volleyball-news", "Волейбол", "Новости волейбола на Sports.ru.", "https://www.sports.ru/rss/rubric/214.xml", []string{"volleyball", "спорт", "волейбол"}),
				sportsSource("sports-athletics-news", "Легкая атлетика", "Новости легкой атлетики на Sports.ru.", "https://www.sports.ru/rss/rubric/215.xml", []string{"athletics", "спорт", "легкая атлетика"}),
				sportsSource("sports-cycling-news", "Велоспорт", "Новости велоспорта на Sports.ru.", "https://www.sports.ru/rss/rubric/216.xml", []string{"cycling", "спорт", "велоспорт"}),
				sportsSource("sports-water-sports-news", "Водные виды", "Новости водных видов спорта на Sports.ru.", "https://www.sports.ru/rss/rubric/217.xml", []string{"water sports", "спорт", "плавание"}),
				sportsSource("sports-chess-news", "Шахматы", "Новости шахмат на Sports.ru.", "https://www.sports.ru/rss/rubric/218.xml", []string{"chess", "спорт", "шахматы"}),
				sportsSource("sports-futsal-news", "Футзал", "Новости футзала на Sports.ru.", "https://www.sports.ru/rss/rubric/219.xml", []string{"futsal", "спорт", "футзал"}),
				sportsSource("sports-handball-news", "Гандбол", "Новости гандбола на Sports.ru.", "https://www.sports.ru/rss/rubric/220.xml", []string{"handball", "спорт", "гандбол"}),
				sportsSource("sports-gymnastics-news", "Гимнастика", "Новости гимнастики на Sports.ru.", "https://www.sports.ru/rss/rubric/221.xml", []string{"gymnastics", "спорт", "гимнастика"}),
				sportsSource("sports-figure-skating-news", "Фигурное катание", "Новости фигурного катания на Sports.ru.", "https://www.sports.ru/rss/rubric/223.xml", []string{"figure skating", "спорт", "фигурное катание"}),
				sportsSource("sports-biathlon-news", "Биатлон", "Новости биатлона на Sports.ru.", "https://www.sports.ru/rss/rubric/225.xml", []string{"biathlon", "спорт", "биатлон"}),
			},
		},
	}
}

func FindCatalogSource(id string) (CatalogSource, bool) {
	for _, topic := range Topics() {
		for _, source := range topic.Sources {
			if source.ID == id {
				return source, true
			}
		}
	}
	return CatalogSource{}, false
}

func habrSource(id string, title string, description string, pageURL string, tags []string) CatalogSource {
	return CatalogSource{
		ID:          id,
		Provider:    "habr",
		Title:       title,
		Description: description,
		PageURL:     pageURL,
		Tags:        tags,
	}
}

func sportsSource(id string, title string, description string, feedURL string, tags []string) CatalogSource {
	return CatalogSource{
		ID:          id,
		Provider:    "sports",
		Title:       title,
		Description: description,
		PageURL:     feedURL,
		FeedURL:     feedURL,
		Tags:        tags,
	}
}

func ProviderTitle(provider string) string {
	switch provider {
	case "habr":
		return "Habr"
	case "sports":
		return "Sports.ru"
	default:
		return provider
	}
}
