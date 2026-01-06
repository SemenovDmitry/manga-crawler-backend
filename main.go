package main

import (
	"fmt"
	"log"
	"time"

	"github.com/SemenovDmitry/manga-crawler-backend/parsers"
	"github.com/SemenovDmitry/manga-crawler-backend/sites/readmanga"
	"github.com/SemenovDmitry/manga-crawler-backend/storage"
	"github.com/SemenovDmitry/manga-crawler-backend/telegram"
)

func main() {
	tgbot := telegram.InitTelegramBot()

	checkMangaUpdates(tgbot)

	ticker := time.NewTicker(10 * time.Minute)

	defer ticker.Stop()

	log.Println("Manga tracker запущен. Проверка каждую минуту...")

	for range ticker.C {
		checkMangaUpdates(tgbot)
	}
}

func checkMangaUpdates(telegramBot *telegram.TelegramBot) {
	// ParseMgeko()
	readmanga.ReadmangaCrawler(telegramBot)
}

var MangaList = []string{
	"https://www.mgeko.cc/manga/song-of-the-sky-walkers/",
	"https://www.mgeko.cc/manga/surviving-as-a-genius-on-borrowed-time/",
	// "https://www.mgeko.cc/manga/absolute-regression/",
	// "https://www.mgeko.cc/manga/survival-story-of-a-sword-king/",
	// ← Добавляй сюда новые ссылки в таком же формате
}

func ParseMgeko() {
	fmt.Printf("Запуск проверки %d манг на mgeko.cc...\n\n", len(MangaList))

	for i, mangaURL := range MangaList {
		fmt.Printf("[%d/%d] Парсинг: %s\n", i+1, len(MangaList), mangaURL)

		info, err := parsers.MgekoParser(mangaURL)

		if err != nil {
			log.Printf("Ошибка при парсинге: %v\n\n", err)
			continue
		}

		if len(info.Chapters) == 0 {
			fmt.Printf("Главы не найдены: %s\n\n", info.Title)
			continue
		}

		latest := info.Chapters[0] // ← всегда самая новая

		fmt.Printf("Манга: %s\n", info.Title)
		// fmt.Printf("Последняя глава: %s\n", latest.Number)
		fmt.Printf("Ссылка: %s\n\n", latest.URL)

		storage.SaveMangaInfoToJson(info, "mgeko.json")

		// Ждём 4 секунды перед следующей мангой (чтобы не получить бан)
		if i < len(MangaList)-1 {
			fmt.Println("Ожидание 4 секунды...")
			time.Sleep(4 * time.Second)
		}
	}

	fmt.Println("Все манги проверены.")
}
