package identity

type DueDate struct {
	DueDateID    int    `json:"due_date_id"`
	BorrowID     int    `json:"borrow_id"`
	Difference   int    `json:"difference"`
	AmountOfFine string `json:"amount_of_fine"`
	CreatedAt    int    `json:"created_at"`
	UpdatedAt    int    `json:"updated_at"`
}

type DueDateResponse struct {
	Difference   int    `json:"difference"`
	AmountOfFine string `json:"amount_of_fine"`
}
