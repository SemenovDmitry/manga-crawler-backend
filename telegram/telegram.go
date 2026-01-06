package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SemenovDmitry/manga-crawler-backend/types"
	"github.com/joho/godotenv"
)

// Конфигурация Telegram бота
type TelegramBot struct {
	Token   string
	ChatID  string
	Client  *http.Client
	BotURL  string
	Enabled bool
}

// Структура для отправки сообщения
type Message struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
}

// Структура ответа от Telegram API
type APIResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// Инициализация бота из конфига
func InitTelegramBot() *TelegramBot {
	// Загружаем переменные из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	// Получаем переменные
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	var tgBot = newTelegramBot(token, chatID)

	if tgBot.Enabled {
		log.Println("Telegram бот инициализирован")
		// Отправляем уведомление о запуске
		go tgBot.SendStartupNotification()
	} else {
		log.Println("Telegram бот отключен")
	}

	return tgBot
}

// Создание нового бота
func newTelegramBot(token, chatID string) *TelegramBot {
	if token == "" || chatID == "" {
		log.Println("Telegram бот отключен: отсутствуют токен или chatID")
		return &TelegramBot{Enabled: false}
	}

	return &TelegramBot{
		Token:  token,
		ChatID: chatID,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		BotURL:  fmt.Sprintf("https://api.telegram.org/bot%s", token),
		Enabled: true,
	}
}

// Отправка простого текстового сообщения
func (bot *TelegramBot) SendMessage(text string) error {
	if !bot.Enabled {
		log.Println("Telegram бот отключен, сообщение не отправлено")
		return nil
	}

	message := Message{
		ChatID:    bot.ChatID,
		Text:      text,
		ParseMode: "HTML", // Для форматирования HTML
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}

	resp, err := bot.Client.Post(
		fmt.Sprintf("%s/sendMessage", bot.BotURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("ошибка Telegram API: %s", apiResp.Description)
	}

	log.Printf("Telegram сообщение отправлено: %s", text[:min(50, len(text))])
	return nil
}

func (bot *TelegramBot) SendMangaUpdate(manga *types.Manga, newChapters []types.Chapter) error {
	if !bot.Enabled || len(newChapters) == 0 {
		return nil
	}

	var message strings.Builder

	// Заголовок
	message.WriteString(fmt.Sprintf("<b>Манга:</b> %s\n", escapeHTML(manga.Title)))
	message.WriteString(fmt.Sprintf("<b>Ссылка:</b> <a href=\"%s\">Открыть мангу</a>\n\n", manga.Url))

	// Список новых глав
	message.WriteString("<b>Новые главы:</b>\n")
	for i, chapter := range newChapters {
		message.WriteString(fmt.Sprintf("%d. <a href=\"%s\">%s</a>\n",
			i+1, chapter.URL, escapeHTML(chapter.Title)))
	}

	// Отправляем сообщение с ОТКЛЮЧЕННЫМ превью
	return bot.sendMessageWithDisabledPreview(message.String())
}

// Простая функция отправки с отключенным превью
func (bot *TelegramBot) sendMessageWithDisabledPreview(text string) error {
	if !bot.Enabled {
		return nil
	}

	message := Message{
		ChatID:                bot.ChatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true, // ВАЖНО: это отключает Open Graph превью
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}

	resp, err := bot.Client.Post(
		fmt.Sprintf("%s/sendMessage", bot.BotURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("ошибка Telegram API: %s", apiResp.Description)
	}

	log.Printf("Telegram сообщение отправлено (превью отключено)")
	return nil
}

// Уведомление о первой проверке
func (bot *TelegramBot) SendNewMangaNotification(manga *types.Manga) error {
	if !bot.Enabled {
		return nil
	}

	message := fmt.Sprintf(
		"✅ <b>Добавлена новая манга для отслеживания!</b>\n\n"+
			"<b>Название:</b> %s\n"+
			"<b>Ссылка:</b> <a href=\"%s\">Открыть</a>\n"+
			"<b>Количество глав:</b> %d\n\n"+
			"<i>Теперь буду отслеживать обновления этой манги!</i>",
		escapeHTML(manga.Title),
		manga.Url,
		len(manga.Chapters),
	)

	return bot.SendMessage(message)
}

// Уведомление о запуске системы
func (bot *TelegramBot) SendStartupNotification() error {
	if !bot.Enabled {
		return nil
	}

	message := "🚀 <b>Манга-трекер запущен!</b>\n\n" +
		"<i>Начинаю отслеживание обновлений...</i>"

	return bot.SendMessage(message)
}

// Экранирование HTML символов
func escapeHTML(text string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(text)
}
