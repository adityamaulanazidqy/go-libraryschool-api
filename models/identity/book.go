package identity

type Book struct {
	BookID          int    `json:"book_id" validate:"required"`
	Title           string `json:"title" validate:"required"`
	Description     string `json:"description" validate:"required"`
	Author          string `json:"author" validate:"required"`
	Cover           string `json:"cover" validate:"required"`
	Isbn            string `json:"isbn" validate:"required"`
	PublicationYear int    `json:"publication_year" validate:"required"`
	GenreID         int    `json:"genre_id" validate:"required"`
	Genre           string `json:"genre" validate:"required"`
	Quantity        int    `json:"quantity" validate:"required"`
}
