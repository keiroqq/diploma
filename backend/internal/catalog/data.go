package catalog

import "strings"

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
			ID:          "it-technology",
			Title:       "IT и технологии",
			Description: "Разработка, инфраструктура, безопасность, AI, данные, телеком и цифровые сервисы.",
			Sources: []CatalogSource{
				habrSource("habr-backend-news", "Бэкенд", "Новости backend-разработки на Хабре.", "https://habr.com/ru/flows/backend/news/", []string{"backend", "api", "go", "java", "python"}),
				habrSource("habr-frontend-news", "Фронтенд", "Новости frontend-разработки на Хабре.", "https://habr.com/ru/flows/frontend/news/", []string{"frontend", "javascript", "typescript", "react"}),
				habrSource("habr-mobile-news", "Мобильная разработка", "Новости мобильной разработки на Хабре.", "https://habr.com/ru/flows/mobile_development/news/", []string{"mobile", "android", "ios", "flutter"}),
				habrSource("habr-gamedev-news", "GameDev", "Новости разработки игр на Хабре.", "https://habr.com/ru/flows/gamedev/news/", []string{"gamedev", "games", "unity", "unreal"}),
				habrSource("habr-admin-news", "Администрирование", "Новости системного администрирования на Хабре.", "https://habr.com/ru/flows/admin/news/", []string{"admin", "linux", "servers"}),
				habrSource("habr-devops-news", "DevOps", "Новости DevOps и эксплуатации на Хабре.", "https://habr.com/ru/hubs/devops/news/", []string{"devops", "docker", "kubernetes", "ci/cd"}),
				habrSource("habr-databases-news", "Базы данных", "Новости баз данных и хранения данных на Хабре.", "https://habr.com/ru/hubs/databases/news/", []string{"databases", "postgresql", "sql"}),
				habrSource("habr-analytics-news", "Аналитика данных", "Новости аналитики и работы с данными на Хабре.", "https://habr.com/ru/flows/analytics/news/", []string{"analytics", "data", "bi"}),
				habrSource("habr-security-news", "Информационная безопасность", "Новости ИБ на Хабре.", "https://habr.com/ru/flows/information_security/news/", []string{"security", "infosec", "pentest"}),
				habrSource("habr-ai-ml-news", "AI и ML", "Новости искусственного интеллекта и машинного обучения на Хабре.", "https://habr.com/ru/flows/ai_and_ml/news/", []string{"ai", "ml", "llm", "нейросети"}),
				vedomostiSource("vedomosti-technology-news", "Технологии", "Новости технологий от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/technology.xml", []string{"technology", "технологии"}),
				kommersantSource("kommersant-telecom-news", "Телекоммуникации", "Новости телекоммуникаций от Коммерсанта.", "https://www.kommersant.ru/RSS/section-telecom.xml", []string{"telecom", "technology", "телекоммуникации"}),
				kommersantSource("kommersant-hitech-news", "Hi-tech", "Новости hi-tech от Коммерсанта.", "https://www.kommersant.ru/RSS/section-hitech.xml", []string{"technology", "hi-tech", "технологии"}),
			},
		},
		{
			ID:          "business-economics",
			Title:       "Бизнес, экономика и финансы",
			Description: "Компании, рынки, финансы, экономика, менеджмент, маркетинг и потребительский рынок.",
			Sources: []CatalogSource{
				habrSource("habr-management-news", "Менеджмент", "Новости управления проектами и командами на Хабре.", "https://habr.com/ru/flows/management/news/", []string{"management", "product", "team"}),
				habrSource("habr-marketing-news", "Маркетинг", "Новости IT-маркетинга на Хабре.", "https://habr.com/ru/flows/marketing/news/", []string{"marketing", "growth", "sales"}),
				vedomostiSource("vedomosti-business-news", "Бизнес", "Новости бизнеса от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/business.xml", []string{"business", "бизнес"}),
				vedomostiSource("vedomosti-economics-news", "Экономика", "Новости экономики от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/economics.xml", []string{"economics", "экономика"}),
				vedomostiSource("vedomosti-finance-news", "Финансы", "Новости финансового рынка от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/finance.xml", []string{"finance", "финансы"}),
				vedomostiSource("vedomosti-management-news", "Менеджмент", "Материалы о менеджменте от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/management.xml", []string{"management", "менеджмент"}),
				kommersantSource("kommersant-economics-news", "Экономика", "Новости экономики от Коммерсанта.", "https://www.kommersant.ru/RSS/section-economics.xml", []string{"economics", "экономика"}),
				kommersantSource("kommersant-business-news", "Бизнес", "Новости бизнеса от Коммерсанта.", "https://www.kommersant.ru/RSS/section-business.xml", []string{"business", "бизнес"}),
				kommersantSource("kommersant-market-news", "Потребительский рынок", "Новости потребительского рынка от Коммерсанта.", "https://www.kommersant.ru/RSS/section-market.xml", []string{"market", "business", "потребительский рынок"}),
			},
		},
		{
			ID:          "politics-society-world",
			Title:       "Политика, общество и мир",
			Description: "Политика, международная повестка, общество, мнения и аналитика.",
			Sources: []CatalogSource{
				vedomostiSource("vedomosti-politics-news", "Политика", "Новости политики от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/politics.xml", []string{"politics", "политика"}),
				vedomostiSource("vedomosti-opinion-news", "Мнения", "Колонки и мнения от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/opinion.xml", []string{"opinion", "мнения"}),
				kommersantSource("kommersant-politics-news", "Политика", "Новости политики от Коммерсанта.", "https://www.kommersant.ru/RSS/section-politics.xml", []string{"politics", "политика"}),
				kommersantSource("kommersant-world-news", "В мире", "Международные новости от Коммерсанта.", "https://www.kommersant.ru/RSS/section-world.xml", []string{"world", "мир", "политика"}),
				kommersantSource("kommersant-society-news", "Общество", "Новости общества от Коммерсанта.", "https://www.kommersant.ru/RSS/section-society.xml", []string{"society", "общество"}),
			},
		},
		{
			ID:          "accidents",
			Title:       "Происшествия",
			Description: "ЧП, инциденты, расследования и оперативная хроника.",
			Sources: []CatalogSource{
				kommersantSource("kommersant-accidents-news", "Происшествия", "Новости происшествий от Коммерсанта.", "https://www.kommersant.ru/RSS/section-accidents.xml", []string{"accidents", "происшествия"}),
			},
		},
		{
			ID:          "science-hardware",
			Title:       "Наука и железо",
			Description: "Научпоп, hardware, электроника, инженерные и DIY-темы.",
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
				kommersantSource("kommersant-sport-news", "Спорт", "Спортивные новости от Коммерсанта.", "https://www.kommersant.ru/RSS/section-sport.xml", []string{"sports", "спорт"}),
			},
		},
		{
			ID:          "realty-auto",
			Title:       "Недвижимость и авто",
			Description: "Недвижимость, девелопмент, автомобильный рынок, транспорт и производители.",
			Sources: []CatalogSource{
				vedomostiSource("vedomosti-realty-news", "Недвижимость", "Новости недвижимости от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/realty.xml", []string{"realty", "недвижимость"}),
				vedomostiSource("vedomosti-auto-news", "Авто", "Новости автомобильного рынка от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/auto.xml", []string{"auto", "авто", "автомобили"}),
				kommersantSource("kommersant-auto-news", "Авто", "Новости авто от Коммерсанта.", "https://www.kommersant.ru/RSS/section-auto.xml", []string{"auto", "авто", "автомобили"}),
			},
		},
		{
			ID:          "culture-lifestyle",
			Title:       "Культура и стиль жизни",
			Description: "Культура, стиль, дизайн, лайфстайл и городские практики.",
			Sources: []CatalogSource{
				habrSource("habr-design-news", "Дизайн", "Новости дизайна интерфейсов и продуктов на Хабре.", "https://habr.com/ru/flows/design/news/", []string{"design", "ui", "ux"}),
				vedomostiSource("vedomosti-lifestyle-news", "Стиль жизни", "Материалы о стиле жизни от Ведомостей.", "https://www.vedomosti.ru/rss/rubric/lifestyle.xml", []string{"lifestyle", "стиль жизни"}),
				kommersantSource("kommersant-culture-news", "Культура", "Новости культуры от Коммерсанта.", "https://www.kommersant.ru/RSS/section-culture.xml", []string{"culture", "культура"}),
				kommersantSource("kommersant-style-news", "Стиль", "Материалы о стиле от Коммерсанта.", "https://www.kommersant.ru/RSS/section-style.xml", []string{"lifestyle", "стиль жизни"}),
			},
		},
		{
			ID:          "regions",
			Title:       "Регионы",
			Description: "Региональная повестка крупных городов и деловых центров.",
			Sources: []CatalogSource{
				kommersantSource("kommersant-spb-news", "Санкт-Петербург", "Региональные новости Коммерсанта по Санкт-Петербургу.", "https://www.kommersant.ru/rss/regions/piter_all.xml", []string{"regions", "санкт-петербург"}),
				kommersantSource("kommersant-ekaterinburg-news", "Екатеринбург", "Региональные новости Коммерсанта по Екатеринбургу.", "https://www.kommersant.ru/rss/regions/ekaterinburg_all.xml", []string{"regions", "екатеринбург"}),
				kommersantSource("kommersant-novosibirsk-news", "Новосибирск", "Региональные новости Коммерсанта по Новосибирску.", "https://www.kommersant.ru/rss/regions/novosibirsk_all.xml", []string{"regions", "новосибирск"}),
				kommersantSource("kommersant-samara-news", "Самара", "Региональные новости Коммерсанта по Самаре.", "https://www.kommersant.ru/rss/regions/samara_all.xml", []string{"regions", "самара"}),
				kommersantSource("kommersant-kazan-news", "Казань", "Региональные новости Коммерсанта по Казани.", "https://www.kommersant.ru/rss/regions/kazan_all.xml", []string{"regions", "казань"}),
				kommersantSource("kommersant-krasnodar-news", "Краснодар", "Региональные новости Коммерсанта по Краснодару.", "https://www.kommersant.ru/rss/regions/krasnodar_all.xml", []string{"regions", "краснодар"}),
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
		FeedURL:     habrFeedURL(pageURL),
		Tags:        tags,
	}
}

func habrFeedURL(pageURL string) string {
	pageURL = strings.TrimSpace(pageURL)
	if pageURL == "" {
		return ""
	}

	feedURL := pageURL
	if !strings.Contains(feedURL, "/ru/rss/") {
		feedURL = strings.Replace(feedURL, "/ru/", "/ru/rss/", 1)
	}
	if !strings.Contains(feedURL, "?") {
		feedURL += "?fl=ru"
	}
	return feedURL
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

func vedomostiSource(id string, title string, description string, feedURL string, tags []string) CatalogSource {
	return CatalogSource{
		ID:          id,
		Provider:    "vedomosti",
		Title:       title,
		Description: description,
		PageURL:     feedURL,
		FeedURL:     feedURL,
		Tags:        tags,
	}
}

func kommersantSource(id string, title string, description string, feedURL string, tags []string) CatalogSource {
	return CatalogSource{
		ID:          id,
		Provider:    "kommersant",
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
	case "vedomosti":
		return "Ведомости"
	case "kommersant":
		return "Коммерсантъ"
	default:
		return provider
	}
}
