package readmanga

import (
	"fmt"
	"log"
	"time"

	"github.com/SemenovDmitry/manga-crawler-backend/storage"
	"github.com/SemenovDmitry/manga-crawler-backend/utils"
)

const baseUrl = "https://a.zazaza.me"

var mangaList = []string{
	"kak_vyjit_geniiu__ogranichennomu_vo_vremeni",
	"absoliutnaia_regressiia",
	"gatiakuta",
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

		fmt.Printf("Найдено глав: %d\n\n", len(feed.Items))

		transformedFeed, err := utils.TransformRSSFeed(feed)
		if err != nil {
			log.Printf("Ошибка преобразования RSS для %s: %v", mangaName, err)
			break
		}

		savedManga, err := storage.GetMangaByName(transformedFeed.Title, "readmanga.json")

		if err != nil {
			log.Printf("Ошибка: %v", err)
			break
		}

		if savedManga == nil {
			storage.SaveMangaInfoToJson(transformedFeed, "readmanga.json")
			continue
		}

		diff := utils.ChaptersDiff(transformedFeed.Chapters, savedManga.Chapters)

		if diff != nil {
			log.Printf("telegram notify diff: %v", diff)
		}

		// Ждём 4 секунды перед следующей мангой (чтобы не получить бан)
		if index < len(mangaList)-1 {
			fmt.Println("Ожидание 4 секунды...")
			time.Sleep(4 * time.Second)
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
