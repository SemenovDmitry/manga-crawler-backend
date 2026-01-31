package main

import (
	"log"
	"time"

	"github.com/SemenovDmitry/manga-crawler-backend/sites/readmanga"
	"github.com/SemenovDmitry/manga-crawler-backend/telegram"
)

func main() {
	tgbot := telegram.InitTelegramBot()

	checkMangaUpdates(tgbot)

	ticker := time.NewTicker(15 * time.Minute)

	defer ticker.Stop()

	log.Println("Manga tracker запущен. Проверка каждst 15 минут...")

	for range ticker.C {
		checkMangaUpdates(tgbot)
	}
}

func checkMangaUpdates(telegramBot *telegram.TelegramBot) {
	readmanga.ReadmangaCrawler(telegramBot)
}
