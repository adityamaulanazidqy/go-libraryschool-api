package identity

type Book struct {
	BookID          int    `json:"book_id"`
	Title           string `json:"title"`
	Author          string `json:"author"`
	Cover           string `json:"cover"`
	Isbn            string `json:"isbn"`
	PublicationYear int    `json:"publication_year"`
	GenreID         int    `json:"genre_id"`
	Genre           string `json:"genre"`
	Quantity        int    `json:"quantity"`
}
