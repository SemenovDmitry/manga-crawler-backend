package readmanga

import (
	"fmt"
	"log"
	"time"

	"github.com/SemenovDmitry/manga-crawler-backend/storage"
	"github.com/SemenovDmitry/manga-crawler-backend/telegram"
	"github.com/SemenovDmitry/manga-crawler-backend/types"
	"github.com/SemenovDmitry/manga-crawler-backend/utils"
)

const readmangaUrl = "https://a.zazaza.me"
const mintUrl = "https://1.seimanga.me"

var readmangaList = []string{
	"absoliutnaia_regressiia",
	"asura_2024",
	"bashnia_podzemelia",
	"gatiakuta",
	"da__ia_pauk__i_chto_s_togo___A5664",
	"kak_vyjit_geniiu__ogranichennomu_vo_vremeni",
	"kohai_iz_zatvornicy_prevratilas_v_infliuensera",
	"pesn_o_nebesnyh_strannikah__A5238",
	"praviachii_mirom",
	"ugroza_v_moem_serdce__A5238",
	"friren__provojaiuchaia_v_poslednii_put__A533b",
	"reinkarnaciia_nebesnogo_demona",
	"asura_2024",
}

var mintList = []string{
	"ad_58",
	"dandadan",
	"devochka__kotoraia_vidit_eto",
	"drama_kvin",
	"ubiica_goblinov",
	"ubiica_goblinov__god_pervyi",
	"carstvo",
	"chelovek_benzopila_2",
	"diavolskii_ostrov",
}

func ParseSourse(telegramBot *telegram.TelegramBot, sourceUrl string, checkList []string) error {
	log.Printf("Начало проверки обновлений манги...\n\n")

	for index, mangaName := range checkList {
		log.Printf("Проверяем мангу: %s", mangaName)

		rssUrl, err := findRSSLink(sourceUrl, mangaName)
		if err != nil {
			log.Printf("Ошибка поиска RSS ссылки для %s: %v", mangaName, err)
			telegramBot.SendErrorNotification(mangaName)
			continue
		}

		feed, err := getRSSFeed(rssUrl)

		if err != nil {
			log.Printf("Ошибка получения RSS для %s: %v", mangaName, err)
			telegramBot.SendErrorNotification(mangaName)
			continue
		}

		if len(feed.Items) == 0 {
			log.Printf("Нет глав для %s", mangaName)
			continue
		}

		// fmt.Printf("Найдено глав: %d\n\n", len(feed.Items))

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

		storage.SaveMangaInfoToJson(transformedFeed, "readmanga.json")

		var diff []types.Chapter

		if savedManga != nil {
			diff = utils.ChaptersDiff(transformedFeed.Chapters, savedManga.Chapters)
		} else {
			diff = transformedFeed.Chapters
		}

		fmt.Printf("Новых глав: %d\n\n", len(diff))

		if diff != nil {
			telegramBot.SendMangaUpdate(sourceUrl, transformedFeed, diff)
		}

		// Ждём 4 секунды перед следующей мангой (чтобы не получить бан)
		if index < len(checkList)-1 {
			fmt.Println("Ожидание 4 секунды...")
			time.Sleep(4 * time.Second)
		}
	}

	log.Println("Проверка обновлений завершена")
	return nil
}

func ReadmangaCrawler(telegramBot *telegram.TelegramBot) error {
	ParseSourse(telegramBot, readmangaUrl, readmangaList)
	ParseSourse(telegramBot, mintUrl, mintList)

	return nil
}
