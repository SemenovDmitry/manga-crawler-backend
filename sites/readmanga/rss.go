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

	"github.com/SemenovDmitry/manga-crawler-backend/types"
	"github.com/SemenovDmitry/manga-crawler-backend/utils"
)

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

func getRSSFeed(mangaName string) (types.Channel, error) {
	var channel types.Channel
	url := fmt.Sprintf("%s/rss/manga?name=%s", baseUrl, mangaName)

	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          10,
			IdleConnTimeout:       60 * time.Second,
			DisableCompression:    false,
			DisableKeepAlives:     false,
			MaxIdleConnsPerHost:   10,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 5 * time.Second,
		},
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
	var rss types.RSS
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
