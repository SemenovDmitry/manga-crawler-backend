package parsers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SemenovDmitry/manga-crawler-backend/types"
	"github.com/SemenovDmitry/manga-crawler-backend/utils"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

func parseWithColly(url string) (*types.Manga, error) {
	crawler := colly.NewCollector(
		colly.UserAgent(utils.GetRandomUserAgent()),
		colly.MaxDepth(1),
	)

	// ← ВАЖНО: эти 3 колбэка покажут ВСЁ
	crawler.OnRequest(func(r *colly.Request) {
		fmt.Printf("Запрос → %s | UA: %s\n", r.URL.String(), r.Headers.Get("User-Agent"))
	})

	crawler.OnResponse(func(r *colly.Response) {
		fmt.Printf("Ответ ← %d %s | Длина: %d байт\n", r.StatusCode, r.Request.URL, len(r.Body))
		if r.StatusCode == 403 {
			fmt.Println("БЛОКИРОВКА 403! Тебя забанили на этом этапе.")
			// Попробуй открыть этот URL в браузере — увидишь Cloudflare или капчу
		}
	})

	crawler.OnError(func(r *colly.Response, err error) {
		fmt.Printf("ОШИБКА %d → %s | %v\n", r.StatusCode, r.Request.URL, err)
		if r.StatusCode == 403 {
			fmt.Println("ОШИБКА 403 — ЗАБЛОКИРОВАЛИ!")
		}
	})

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

func parseWithChromedp(url string) (*types.Manga, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	var title string
	var chapters []types.Chapter

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`ul.chapter-list`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil),
		chromedp.Sleep(3*time.Second),

		// Получаем название
		chromedp.Text(`h1[itemprop="name"].novel-title`, &title, chromedp.ByQuery, chromedp.NodeVisible),

		// Главная магия: собираем все li → по одному
		chromedp.Tasks{
			chromedp.Evaluate(`
				Array.from(document.querySelectorAll('ul.chapter-list li.chapter-list-item')).map(el => ({
					href: el.querySelector('a')?.getAttribute('href') || '',
					number: el.querySelector('.chapter-number')?.textContent || ''
				}))
			`, &chapters),
		},
	)
	if err != nil {
		return nil, err
	}

	// Обрабатываем результат из JS
	var processed []types.Chapter
	for _, ch := range chapters {
		number := strings.TrimSpace(strings.Split(ch.Number, "\n")[0])
		if number == "" {
			continue
		}

		link := ch.URL
		if strings.HasPrefix(link, "/") {
			link = "https://www.mgeko.cc" + link
		}

		processed = append(processed, types.Chapter{
			Number: number,
			URL:    link,
		})
	}

	if len(processed) == 0 {
		return nil, nil
	}

	return &types.Manga{
		Title:    strings.TrimSpace(title),
		Url:      url,
		Chapters: processed,
	}, nil
}

func MgekoParser(url string) (*types.Manga, error) {
	info, err := parseWithColly(url)

	if err == nil {
		fmt.Printf("Успех через Colly! %s → %s\n", info.Title, info.Chapters[0].Number)
		return info, nil
	}

	// Если Colly не сработал — включаем тяжёлую артиллерию
	fmt.Printf("Colly не справился (%v) → переключаемся на chromedp...\n", err)

	info, err = parseWithChromedp(url)
	if err != nil {
		return nil, fmt.Errorf("оба парсера упали: %w", err)
	}

	fmt.Printf("Успех через chromedp! %s → %s\n", info.Title, info.Chapters[0].Number)
	return info, nil
}
