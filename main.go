package main

import (
	"log"
	"time"

	"github.com/SemenovDmitry/manga-crawler-backend/parsers"
	"github.com/SemenovDmitry/manga-crawler-backend/storage"
	"github.com/SemenovDmitry/manga-crawler-backend/telegram"
)

func main() {
	// Инициализируем подключение к БД
	_ = storage.GetDB()
	defer storage.CloseDB()

	// Запускаем миграции
	if err := storage.RunMigrations(); err != nil {
		log.Fatalf("Ошибка миграций: %v", err)
	}

	// Инициализируем Telegram бота
	tgbot := telegram.InitTelegramBot()

	// Запускаем polling для получения команд от пользователей
	go telegram.StartPolling(tgbot)

	// Запускаем первую проверку обновлений
	checkMangaUpdates(tgbot)

	// Запускаем периодическую проверку каждые 20 минут
	ticker := time.NewTicker(20 * time.Minute)
	defer ticker.Stop()

	log.Println("Manga tracker запущен. Проверка каждые 20 минут...")

	for range ticker.C {
		checkMangaUpdates(tgbot)
	}
}

func checkMangaUpdates(telegramBot *telegram.TelegramBot) {
	sources, err := storage.GetActiveSources()
	if err != nil {
		log.Printf("Ошибка получения источников: %v", err)
		return
	}

	if len(sources) == 0 {
		log.Println("Нет активных источников для парсинга")
		return
	}

	log.Printf("Найдено %d активных источников", len(sources))

	for _, source := range sources {
		mangaList, err := storage.GetMangaBySourceID(source.ID)
		if err != nil {
			log.Printf("Ошибка получения манги для %s: %v", source.ParserName, err)
			continue
		}

		if len(mangaList) == 0 {
			log.Printf("Нет манги для отслеживания в источнике %s", source.ParserName)
			continue
		}

		if err := parsers.RunParser(telegramBot, source, mangaList); err != nil {
			log.Printf("Ошибка парсинга %s: %v", source.ParserName, err)
		}
	}
}
