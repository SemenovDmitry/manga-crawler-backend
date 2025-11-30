package parsers

import (
	"strings"
	"time"

	"github.com/SemenovDmitry/manga-crawler-backend/types"
	"github.com/SemenovDmitry/manga-crawler-backend/utils"
	"github.com/gocolly/colly/v2"
)

func ParseMgekoManga(url string) (*types.Manga, error) {
	crawler := colly.NewCollector(
		colly.UserAgent(utils.GetRandomUserAgent()),
		colly.MaxDepth(1),
	)

	info := &types.Manga{
		Url:      url,
		Chapters: make([]types.Chapter, 0, 50),
	}

	crawler.OnHTML(`h1[itemprop="name"].novel-title`, func(element *colly.HTMLElement) {
		info.Title = strings.TrimSpace(element.Text)
	})

	crawler.OnHTML("ul.chapter-list li.chapter-list-item", func(element *colly.HTMLElement) {
		// Ссылка на главу
		link := element.ChildAttr("a", "href")

		if link == "" {
			return
		}

		if strings.HasPrefix(link, "/") {
			link = "https://www.mgeko.cc" + link
		}

		// Чистый номер главы (только первая строка из .chapter-number)
		raw := element.ChildText(".chapter-number")
		number := strings.Split(strings.TrimSpace(raw), "\n")[0]
		number = strings.TrimSpace(number)

		// Пропускаем пустые или битые элементы
		if number == "" {
			return
		}

		chapter := types.Chapter{
			Number: number,
			URL:    link,
		}

		// fmt.Printf("chapter: %+v\n", chapter)

		// Добавляем в начало — потому что список идёт от новой к старой
		info.Chapters = append(info.Chapters, chapter)
	})

	// Задержка — mgeko.cc чувствителен к частым запросам
	crawler.Limit(&colly.LimitRule{
		DomainGlob:  "*.mgeko.cc",
		Delay:       2 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	err := crawler.Visit(url)

	if err != nil {
		return nil, err
	}

	// fmt.Printf("Успешно спаршено %d глав!\n", len(info.Chapters))

	return info, nil
}
