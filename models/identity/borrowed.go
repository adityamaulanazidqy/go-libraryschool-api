package identity

type Borrowed struct {
	BorrowedID int    `json:"borrowed_id"`
	BookID     int    `json:"book_id"`
	UserID     int    `json:"user_id"`
	BorrowDate string `json:"borrow_date"`
	ReturnDate string `json:"return_date"`
}

type BookBorrowingData struct {
	Borrowed        Borrowed        `json:"borrowed"`
	DueDateResponse DueDateResponse `json:"due_date_response"`
}
