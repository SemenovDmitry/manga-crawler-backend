package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/SemenovDmitry/manga-crawler-backend/types"
)

var (
	mangasDBMutex sync.RWMutex // защита от одновременной записи (если будет cron)
)

func SaveMangaInfoToJson(manga *types.Manga, name string) {
	if manga == nil || manga.Title == "" {
		log.Println("Ошибка: manga или Title пустой — не сохраняем")
		return
	}

	var dbDir = "mangasdb"
	var dbFile = name

	filePath := filepath.Join(dbDir, dbFile)

	// Создаём папку, если нет
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Printf("Не удалось создать папку %s: %v", dbDir, err)
		return
	}

	// Блокируем на время чтения/записи
	mangasDBMutex.Lock()
	defer mangasDBMutex.Unlock()

	// Читаем существующий файл (если есть)
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Ошибка чтения файла %s: %v", filePath, err)
		return
	}

	// Если файла нет — создаём пустой массив
	var mangas []types.Manga
	if len(data) > 0 {
		if jsonErr := json.Unmarshal(data, &mangas); jsonErr != nil {
			log.Printf("Ошибка парсинга JSON: %v — будет создан новый файл", jsonErr)
			mangas = []types.Manga{}
		}
	}

	// Ищем, есть ли уже манга с таким Title
	found := false
	for i := range mangas {
		if mangas[i].Title == manga.Title {
			mangas[i] = *manga // обновляем существующую
			found = true
			break
		}
	}

	// Если не нашли — добавляем новую
	if !found {
		mangas = append(mangas, *manga)
	}

	// Сохраняем обратно
	jsonData, err := json.MarshalIndent(mangas, "", "  ")
	if err != nil {
		log.Printf("Ошибка при создании JSON: %v", err)
		return
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Printf("Ошибка записи файла %s: %v", filePath, err)
		return
	}

	// Успешно!
	latest := "нет глав"
	if len(manga.Chapters) > 0 {
		latest = manga.Chapters[0].Number
	}

	fmt.Printf("Обновлено в БД: %s → глава %s\n", manga.Title, latest)
}
