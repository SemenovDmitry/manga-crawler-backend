package readmanga

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/SemenovDmitry/manga-crawler-backend/storage"
	"github.com/SemenovDmitry/manga-crawler-backend/types"
	"github.com/SemenovDmitry/manga-crawler-backend/utils"
)

const baseUrl = "https://a.zazaza.me"

var mangaList = []string{
	"kak_vyjit_geniiu__ogranichennomu_vo_vremeni",
	"absoliutnaia_regressiia",
	"gatiakuta",
}

// RSS структуры
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Items []Item `xml:"item"`
}

type Item struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

// Функция для декомпрессии gzip
func decompressBody(body []byte) ([]byte, error) {
	// Пробуем распаковать как gzip
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		// Если не gzip, возвращаем как есть
		return body, nil
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return body, nil
	}

	return decompressed, nil
}

func getRSSHeaders() http.Header {
	return http.Header{
		"User-Agent":      {utils.GetRandomUserAgent()},
		"Accept":          {"application/xml,text/xml,application/rss+xml"},
		"Accept-Language": {"ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7"},
		"Accept-Encoding": {"gzip, deflate"}, // Разрешаем сжатие
		"Referer":         {baseUrl + "/"},
		"Connection":      {"keep-alive"},
	}
}

func getRSSFeed(mangaName string) (Channel, error) {
	var channel Channel
	url := fmt.Sprintf("%s/rss/manga?name=%s", baseUrl, mangaName)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return channel, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Важные заголовки
	req.Header = getRSSHeaders()

	resp, err := client.Do(req)
	if err != nil {
		return channel, fmt.Errorf("ошибка запроса RSS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return channel, fmt.Errorf("статус ошибки: %d %s", resp.StatusCode, resp.Status)
	}

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return channel, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// Декомпрессируем если нужно
	decompressedBody, err := decompressBody(body)
	if err != nil {
		log.Printf("Ошибка декомпрессии: %v", err)
		decompressedBody = body
	}

	// Убираем BOM (Byte Order Mark) если есть
	decompressedBody = bytes.TrimPrefix(decompressedBody, []byte{0xEF, 0xBB, 0xBF})
	decompressedBody = bytes.TrimPrefix(decompressedBody, []byte{0xFE, 0xFF})
	decompressedBody = bytes.TrimPrefix(decompressedBody, []byte{0xFF, 0xFE})

	// Парсим XML
	var rss RSS
	if err := xml.Unmarshal(decompressedBody, &rss); err != nil {
		// Логируем первые 300 байт для отладки
		sample := string(decompressedBody)
		if len(sample) > 300 {
			sample = sample[:300]
		}
		log.Printf("Ошибка парсинга XML: %v", err)
		log.Printf("Начало данных: %s", sample)

		// Пробуем исправить XML - убираем невалидные символы
		if err := xml.Unmarshal(decompressedBody, &rss); err != nil {
			return channel, fmt.Errorf("ошибка парсинга XML после очистки: %v", err)
		}
	}

	return rss.Channel, nil
}

func transformRSSFeed(feed Channel) (*types.Manga, error) {
	manga := &types.Manga{
		Title:    feed.Title,
		Url:      feed.Link,
		Chapters: make([]types.Chapter, len(feed.Items)),
	}

	for i, item := range feed.Items {
		manga.Chapters[i] = types.Chapter{
			Title: item.Title,
			URL:   item.Link,
		}
	}

	return manga, nil
}

func ReadmangaCrawler() error {
	log.Println("Начало проверки обновлений манги...")

	for index, mangaName := range mangaList {
		log.Printf("Проверяем мангу: %s", mangaName)

		feed, err := getRSSFeed(mangaName)
		if err != nil {
			log.Printf("Ошибка получения RSS для %s: %v", mangaName, err)
			continue
		}

		if len(feed.Items) == 0 {
			log.Printf("Нет глав для %s", mangaName)
			continue
		}

		// берем последние 10
		limit := 10
		if len(feed.Items) < limit {
			limit = len(feed.Items)
		}

		fmt.Printf("Найдено глав: %d (показываем последние %d)\n\n", len(feed.Items), limit)

		transformedData, err := transformRSSFeed(feed)
		if err != nil {
			log.Printf("Ошибка преобразования RSS для %s: %v", mangaName, err)
			break
		}

		storage.SaveMangaInfoToJson(transformedData, "mgeko.json")

		// Ждём 3 секунды перед следующей мангой (чтобы не получить бан)
		if index < len(mangaList)-1 {
			fmt.Println("Ожидание 3 секунды...")
			time.Sleep(3 * time.Second)
		}

		// for i := 0; i < limit; i++ {
		// 	item := feed.Items[i]

		// 	fmt.Printf("Заголовок: %s\n", strings.TrimSpace(item.Title))
		// 	fmt.Printf("  Ссылка: %s\n\n", item.Link)
		// }

	}

	log.Println("Проверка обновлений завершена")
	return nil
}
