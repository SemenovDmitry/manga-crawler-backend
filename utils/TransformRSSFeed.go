package utils

import "github.com/SemenovDmitry/manga-crawler-backend/types"

func TransformRSSFeed(feed types.Channel) (*types.Manga, error) {
	manga := &types.Manga{
		Title:    feed.Title,
		Url:      feed.Link,
		Chapters: make([]types.Chapter, 10),
	}

	if len(feed.Items) == 0 {
		return manga, nil
	}

	filteredItems := feed.Items[:min(len(feed.Items), 10)]

	for i, item := range filteredItems {
		manga.Chapters[i] = types.Chapter{
			Title: item.Title,
			URL:   item.Link,
		}
	}

	return manga, nil
}
