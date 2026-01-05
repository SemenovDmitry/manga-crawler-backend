package types

type Chapter struct {
	Title string
	URL   string
}

type Manga struct {
	Title    string
	Url      string
	Chapters []Chapter
}
