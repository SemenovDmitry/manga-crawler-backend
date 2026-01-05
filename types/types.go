package types

import "encoding/xml"

type Chapter struct {
	Title string
	URL   string
}

type Manga struct {
	Title    string
	Url      string
	Chapters []Chapter
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
