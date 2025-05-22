package request_models

type BookById struct {
	BookID int `json:"book_id"`
}

type BookByGenreId struct {
	GenreID int `json:"genre_id"`
}

type BookByTitle struct {
	Title string `json:"title"`
}

type FavoriteBookRequest struct {
	UserID int `json:"user_id"`
	BookID int `json:"book_id"`
}

type BookUpdate struct {
	BookID      int    `json:"book_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`
	Cover       string `json:"cover,omitempty"`
	GenreID     int    `json:"genre_id,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Quantity    int    `json:"quantity,omitempty"`
}

type BookBorrowedRequest struct {
	BookID       int    `json:"book_id"`
	UserID       int    `json:"user_id"`
	BorrowedDate string `json:"borrowed_date"`
	ReturnDate   string `json:"return_date"`
}

type HistoryRecordingBorrowed struct {
	UserID     int `json:"user_id"`
	BorrowedID int `json:"borrowed_id"`
}
