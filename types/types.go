package types

type Chapter struct {
	Number string
	Title  string
	URL    string
}

type Manga struct {
	Title    string
	Url      string
	Chapters []Chapter
}
