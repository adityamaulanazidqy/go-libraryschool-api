package management_book_repository

import (
	context2 "context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/helpers"
	"go-libraryschool/models/identity"
	"go-libraryschool/models/request_models"
	"net/http"
	"time"
)

type ManagementBookRepository struct {
	db        *sql.DB
	logLogrus *logrus.Logger
	rdb       *redis.Client
}

func NewManagementBookRepository(db *sql.DB, logLogrus *logrus.Logger, rdb *redis.Client) *ManagementBookRepository {
	return &ManagementBookRepository{db: db, logLogrus: logLogrus, rdb: rdb}
}

func (repository *ManagementBookRepository) AddBookRepository(r *http.Request, book *identity.Book) (helpers.ApiResponse, int, error) {
	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "INSERT INTO books (title, author, cover, genre_id, isbn, publication_year, quantity) VALUES (?, ?, ?, ?, ?, ?, ?)"

	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to prepare statement")

		result := helpers.ApiResponse{
			Message: "Failed to prepare statement",
			Data:    nil,
		}

		return result, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, book.Title, book.Author, book.Cover, book.GenreID, book.Isbn, book.PublicationYear, book.Quantity)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to execute statement")

		resultExec := helpers.ApiResponse{
			Message: "Failed to execute statement",
			Data:    nil,
		}

		return resultExec, http.StatusInternalServerError, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to get rows affected")

		resultRowsAffected := helpers.ApiResponse{
			Message: "Failed to get rows affected",
			Data:    nil,
		}

		return resultRowsAffected, http.StatusInternalServerError, err
	}

	repository.logLogrus.Info("Rows affected: ", rowsAffected)

	queryGenre := "SELECT genre_name FROM genres WHERE genre_id = ?"
	stmt, err = repository.db.PrepareContext(ctx, queryGenre)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement"}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, book.GenreID).Scan(&book.Genre)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute statement",
		}).Error("Failed to execute statement")

		return helpers.ApiResponse{Message: "Failed to execute statement"}, http.StatusInternalServerError, err
	}

	data := map[string]any{
		"title":           book.Title,
		"author":          book.Author,
		"cover":           book.Cover,
		"genre":           book.Genre,
		"isbn":            book.Isbn,
		"publicationYear": book.PublicationYear,
		"quantity":        book.Quantity,
	}

	resultLate := helpers.ApiResponse{
		Message: "Successfully added book",
		Data:    data,
	}

	return resultLate, http.StatusCreated, nil
}

func (repository *ManagementBookRepository) GetBooksRepository() (helpers.ApiResponse, int, error) {
	var books []identity.Book

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	booksRedis, err := repository.rdb.Get(ctx, "Books").Result()
	if err != nil {
		responseRepo, code, err := repository.GetBooksDBMysql()
		return responseRepo, code, err
	}

	err = json.Unmarshal([]byte(booksRedis), &books)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  booksRedis,
		}).Error("Failed to unmarshal books from Redis")

		result := helpers.ApiResponse{
			Message: "Failed to decode books from Redis",
			Data:    nil,
		}
		return result, http.StatusInternalServerError, err
	}

	result := helpers.ApiResponse{
		Message: "Successfully retrieved books from Redis",
		Data:    books,
	}
	return result, http.StatusOK, nil
}

func (repository *ManagementBookRepository) GetBooksDBMysql() (helpers.ApiResponse, int, error) {
	var books []identity.Book

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT book_id, title, author, cover, genre_id, isbn, publication_year, quantity FROM books"
	rows, err := repository.db.QueryContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{"error": err}).Error("Failed to execute query books")
		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}
	defer rows.Close()

	for rows.Next() {
		var book identity.Book
		err = rows.Scan(&book.BookID, &book.Title, &book.Author, &book.Cover, &book.GenreID, &book.Isbn, &book.PublicationYear, &book.Quantity)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{"error": err}).Error("Failed to scan book row")
			return helpers.ApiResponse{Message: "Failed to scan book row", Data: nil}, http.StatusInternalServerError, err
		}

		err = repository.db.QueryRowContext(ctx, "SELECT genre_name FROM genres WHERE genre_id = ?", book.GenreID).Scan(&book.Genre)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get genre name")
			return helpers.ApiResponse{Message: "Failed to get genre name", Data: nil}, http.StatusInternalServerError, err
		}

		books = append(books, book)
	}

	if len(books) == 0 {
		repository.logLogrus.Info("No books found")
		return helpers.ApiResponse{Message: "No books found", Data: nil}, http.StatusNotFound, nil
	}

	booksJSON, err := json.Marshal(books)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{"error": err}).Error("Failed to marshal books")
		return helpers.ApiResponse{Message: "Failed to marshal books", Data: nil}, http.StatusInternalServerError, err
	}

	err = repository.rdb.Set(ctx, "Books", booksJSON, 30*time.Second).Err()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{"error": err}).Error("Failed to set books into Redis")
		return helpers.ApiResponse{Message: "Failed to set books into Redis", Data: nil}, http.StatusInternalServerError, err
	}

	repository.logLogrus.Info("Successfully cached books in Redis")

	result := helpers.ApiResponse{
		Message: "Successfully retrieved books from database",
		Data:    books,
	}
	return result, http.StatusOK, nil
}

func (repository *ManagementBookRepository) GetBookRepository(r *http.Request, bookID int) (helpers.ApiResponse, int, error) {
	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	var book identity.Book

	query := "SELECT book_id, title, author, cover, genre_id, isbn, publication_year, quantity FROM books WHERE book_id = ?"
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to prepare statement")

		result := helpers.ApiResponse{
			Message: "Failed to prepare statement",
			Data:    nil,
		}
		return result, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, bookID).Scan(&book.BookID, &book.Title, &book.Author, &book.Cover, &book.GenreID, &book.Isbn, &book.PublicationYear, &book.Quantity)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error": err,
				"data":  r.Body,
			}).Info("No book found")

			result := helpers.ApiResponse{
				Message: "No book found",
				Data:    nil,
			}
			return result, http.StatusNotFound, err
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to query row context")

		result := helpers.ApiResponse{
			Message: "Failed to query row context",
			Data:    nil,
		}
		return result, http.StatusInternalServerError, err
	}

	queryGenre := "SELECT genre_name FROM genres WHERE genre_id = ?"
	stmt, err = repository.db.PrepareContext(ctx, queryGenre)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to prepare statement")

		result := helpers.ApiResponse{
			Message: "Failed to prepare statement",
			Data:    nil,
		}
		return result, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, book.GenreID).Scan(&book.Genre)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to query row context")

		result := helpers.ApiResponse{
			Message: "Failed to query row context",
			Data:    nil,
		}
		return result, http.StatusInternalServerError, err
	}

	result := map[string]any{
		"Title":           book.Title,
		"Author":          book.Author,
		"Cover":           book.Cover,
		"Genre":           book.Genre,
		"Isbn":            book.Isbn,
		"PublicationYear": book.PublicationYear,
		"Quantity":        book.Quantity,
	}

	resultLate := helpers.ApiResponse{
		Message: "Successfully retrieved book",
		Data:    result,
	}
	return resultLate, http.StatusOK, nil
}

func (repository *ManagementBookRepository) DeleteBookRepository(r *http.Request, bookID int) (helpers.ApiResponse, int, error) {
	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	var (
		bookIDSelected        int
		bookSelectedForDelete = struct {
			Title  string
			Author string
			Genre  string
		}{}
	)

	querySelectBook := `SELECT title, author, genre_id 
						FROM books 
						WHERE book_id = ?`
	err := repository.db.QueryRowContext(ctx, querySelectBook, bookID).Scan(&bookSelectedForDelete.Title, &bookSelectedForDelete.Author, &bookIDSelected)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to find book")

		resultQuery := helpers.ApiResponse{
			Message: "Failed to find book",
			Data:    nil,
		}
		return resultQuery, http.StatusInternalServerError, err
	}

	querySelectGenre := `SELECT genre_name 
						 FROM genres 
						 WHERE genre_id = ?`
	err = repository.db.QueryRowContext(ctx, querySelectGenre, bookIDSelected).Scan(&bookSelectedForDelete.Genre)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to execute query")

		resultQuery := helpers.ApiResponse{
			Message: "Failed to execute query selected genre book for deletion",
			Data:    nil,
		}
		return resultQuery, http.StatusInternalServerError, err
	}

	queryDelete := "DELETE FROM books WHERE book_id = ?"
	stmt, err := repository.db.PrepareContext(ctx, queryDelete)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to prepare statement")

		resultPrepareDelete := helpers.ApiResponse{
			Message: "Failed to Prepare Context Delete",
			Data:    nil,
		}
		return resultPrepareDelete, http.StatusInternalServerError, err
	}

	result, err := stmt.ExecContext(ctx, bookID)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to execute statement")

		resultExec := helpers.ApiResponse{
			Message: "Failed to execute statement",
			Data:    nil,
		}
		return resultExec, http.StatusInternalServerError, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to get rows affected")

		resultRowAffected := helpers.ApiResponse{
			Message: "Failed to get rows affected",
			Data:    nil,
		}
		return resultRowAffected, http.StatusInternalServerError, err
	}

	repository.logLogrus.Info("Rows affected: ", rowsAffected)

	resultLate := helpers.ApiResponse{
		Message: "Successfully deleted book",
		Data:    bookSelectedForDelete,
	}

	return resultLate, http.StatusOK, nil
}

func (repository *ManagementBookRepository) UpdateBookRepository(r *http.Request, book request_models.BookUpdate) (helpers.ApiResponse, int, error) {
	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	var existingBook request_models.BookUpdate
	query := "SELECT book_id, title, author, cover, genre_id, quantity FROM books WHERE book_id = ?"
	err := repository.db.QueryRowContext(ctx, query, book.BookID).
		Scan(&existingBook.BookID, &existingBook.Title, &existingBook.Author, &existingBook.Cover, &existingBook.GenreID, &existingBook.Quantity)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to select existing book")

		return helpers.ApiResponse{Message: "Failed to find book", Data: nil}, http.StatusInternalServerError, err
	}

	if book.Title == "" {
		book.Title = existingBook.Title
	}
	if book.Author == "" {
		book.Author = existingBook.Author
	}

	if book.Cover == "" {
		book.Cover = existingBook.Cover
	}

	if book.GenreID == 0 {
		book.GenreID = existingBook.GenreID
	}
	if book.Quantity == 0 {
		book.Quantity = existingBook.Quantity
	}

	updateQuery := `
		UPDATE books 
		SET title = ?, author = ?, cover = ?, genre_id = ?, quantity = ?, updated_at = ?
		WHERE book_id = ?`
	stmt, err := repository.db.PrepareContext(ctx, updateQuery)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to prepare update statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := stmt.ExecContext(ctx, book.Title, book.Author, book.Cover, book.GenreID, book.Quantity, now, book.BookID)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to execute update")

		return helpers.ApiResponse{Message: "Failed to execute update", Data: nil}, http.StatusInternalServerError, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to get rows affected")

		return helpers.ApiResponse{Message: "Failed to get rows affected", Data: nil}, http.StatusInternalServerError, err
	}

	if rowsAffected == 0 {
		return helpers.ApiResponse{Message: "No rows updated", Data: nil}, http.StatusNotFound, nil
	}

	repository.logLogrus.WithField("rowsAffected", rowsAffected).Info("Successfully updated book")

	updatedBook := map[string]interface{}{
		"book_id":    book.BookID,
		"title":      book.Title,
		"author":     book.Author,
		"cover":      book.Cover,
		"genre_id":   book.GenreID,
		"quantity":   book.Quantity,
		"updated_at": now,
	}

	return helpers.ApiResponse{Message: "Successfully updated book", Data: updatedBook}, http.StatusOK, nil
}

func (repository *ManagementBookRepository) SearchBooksRepository(r *http.Request, bookTitle string) (helpers.ApiResponse, int, error) {
	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := `SELECT book_id, title, author, cover, genre_id, isbn, publication_year, quantity FROM books WHERE title LIKE ?`
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, "%"+bookTitle+"%")
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}
	defer rows.Close()

	var books []identity.Book

	for rows.Next() {
		var book identity.Book
		err = rows.Scan(&book.BookID, &book.Title, &book.Author, &book.Cover, &book.GenreID, &book.Isbn, &book.PublicationYear, &book.Quantity)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error": err,
				"data":  r.Body,
			}).Error("Failed to scan row")

			return helpers.ApiResponse{Message: "Failed to scan row", Data: nil}, http.StatusInternalServerError, err
		}

		queryGenre := "SELECT genre_name FROM genres WHERE genre_id = ?"
		stmt, err = repository.db.PrepareContext(ctx, queryGenre)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error": err,
				"data":  r.Body,
			}).Error("Failed to prepare statement")

			return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
		}

		err = stmt.QueryRowContext(ctx, book.GenreID).Scan(&book.Genre)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error": err,
				"data":  r.Body,
			}).Error("Failed to execute query")

			return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
		}

		books = append(books, book)
	}

	if len(books) == 0 {
		return helpers.ApiResponse{Message: "No books found", Data: nil}, http.StatusNotFound, nil
	}

	return helpers.ApiResponse{Message: "Successfully searched books", Data: books}, http.StatusOK, nil
}

func (repository *ManagementBookRepository) BorrowedBookRepository(r *http.Request, book request_models.BookBorrowedRequest) (helpers.ApiResponse, int, error) {
	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "INSERT INTO borrowed_books (book_id, user_id, borrow_date, return_date) VALUES (?, ?, ?, ?)"
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, book.BookID, book.UserID, book.BorrowedDate, book.ReturnDate)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to execute statement")

		return helpers.ApiResponse{Message: "Failed to execute statement", Data: nil}, http.StatusInternalServerError, err
	}

	return helpers.ApiResponse{Message: "Successfully borrowed book", Data: book}, http.StatusOK, nil
}

func (repository *ManagementBookRepository) BookBorrowingDataRepository() (helpers.ApiResponse, int, error) {
	var (
		borrowDateStr string
		returnDateStr string
		timeNowStr    = time.Now().Format("2006-01-02")
		layoutDate    = "2006-01-02"
	)
	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT borrow_id, book_id, user_id, borrow_date, return_date FROM borrowed_books"
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute query",
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}
	defer rows.Close()

	var borrowed []identity.BookBorrowingData

	for rows.Next() {
		var borrow identity.BookBorrowingData
		err = rows.Scan(&borrow.Borrowed.BorrowedID, &borrow.Borrowed.BookID, &borrow.Borrowed.UserID, &borrow.Borrowed.BorrowDate, &borrow.Borrowed.ReturnDate)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to scan row",
			}).Error("Failed to scan row")

			return helpers.ApiResponse{Message: "Failed to scan row", Data: nil}, http.StatusInternalServerError, err
		}

		borrowDateStr = borrow.Borrowed.BorrowDate
		returnDateStr = borrow.Borrowed.ReturnDate

		returnDate, err := time.Parse(layoutDate, returnDateStr)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to parse date",
			}).Error("Failed to parse date")

			return helpers.ApiResponse{Message: "Failed to parse date", Data: nil}, http.StatusInternalServerError, err
		}

		timeNow, err := time.Parse(layoutDate, timeNowStr)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to parse time",
			}).Error("Failed to parse time")

			return helpers.ApiResponse{Message: "Failed to parse time", Data: nil}, http.StatusInternalServerError, err
		}

		if timeNow.After(returnDate) {
			difference := timeNow.Sub(returnDate).Hours() / 24
			fineMoneyInt := int(difference) * 20000

			fineMoneyFloat := float64(fineMoneyInt)
			fineMoney := helpers.NewFormatterMoney().FormaterCurrency(fineMoneyFloat, "IDR")

			borrow.DueDateResponse.AmountOfFine = fineMoney
			borrow.DueDateResponse.Difference = int(difference)
			borrow.Borrowed.BorrowDate = borrowDateStr
		}

		borrowed = append(borrowed, borrow)
	}

	return helpers.ApiResponse{Message: "Success", Data: borrowed}, http.StatusOK, nil
}

func (repository *ManagementBookRepository) GetBooksGenreRepository(genreID int) (helpers.ApiResponse, int, error) {
	var books []identity.Book

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT book_id, title, author, cover, genre_id, isbn, publication_year, quantity FROM books WHERE genre_id = ?"
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, genreID)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute query",
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}
	defer rows.Close()

	for rows.Next() {
		var book identity.Book
		err = rows.Scan(&book.BookID, &book.Title, &book.Author, &book.Cover, &book.GenreID, &book.Isbn, &book.PublicationYear, &book.Quantity)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to scan row",
			}).Error("Failed to scan row")

			return helpers.ApiResponse{Message: "Failed to scan row", Data: nil}, http.StatusInternalServerError, err
		}

		queryGenre := "SELECT genre_name FROM genres WHERE genre_id = ?"
		stmt, err = repository.db.PrepareContext(ctx, queryGenre)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to prepare statement",
			}).Error("Failed to prepare statement")

			return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
		}

		err = stmt.QueryRowContext(ctx, book.GenreID).Scan(&book.Genre)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to scan row",
			}).Error("Failed to scan row")

			return helpers.ApiResponse{Message: "Failed to scan row", Data: nil}, http.StatusInternalServerError, err
		}

		books = append(books, book)
	}

	if books == nil {
		return helpers.ApiResponse{Message: "No books found", Data: nil}, http.StatusNotFound, nil
	}

	return helpers.ApiResponse{Message: "Success", Data: books}, http.StatusOK, nil
}

func (repository *ManagementBookRepository) AddFavoriteBookRepository(userID, bookID int) (helpers.ApiResponse, int, error) {
	var (
		user         identity.UserFavoriteBook
		book         identity.Book
		favoriteBook identity.FavoriteBook
	)

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT username, email, roleID FROM users WHERE id = ?"
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, userID).Scan(&user.Username, &user.Email, &user.RoleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "User does not exist",
			}).Error("User does not exist")

			return helpers.ApiResponse{Message: "User does not exist", Data: nil}, http.StatusNotFound, nil
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute query",
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}

	queryRole := "SELECT role FROM roles WHERE id = ?"
	stmt, err = repository.db.PrepareContext(ctx, queryRole)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, user.RoleID).Scan(&user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Role does not exist",
			}).Error("Role does not exist")

			return helpers.ApiResponse{Message: "Role does not exist", Data: nil}, http.StatusNotFound, nil
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute query",
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}

	queryBook := "SELECT book_id, title, author, cover, genre_id, isbn, publication_year, quantity FROM books WHERE book_id = ?"
	stmt, err = repository.db.PrepareContext(ctx, queryBook)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, bookID).Scan(&book.BookID, &book.Title, &book.Author, &book.Cover, &book.GenreID, &book.Isbn, &book.PublicationYear, &book.Quantity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Book does not exist",
			}).Error("Book does not exist")

			return helpers.ApiResponse{Message: "Book does not exist", Data: nil}, http.StatusNotFound, nil
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute query",
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}

	queryGenre := "SELECT genre_name FROM genres WHERE genre_id = ?"
	stmt, err = repository.db.PrepareContext(ctx, queryGenre)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, book.GenreID).Scan(&book.Genre)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Genre does not exist",
			}).Error("Genre does not exist")

			return helpers.ApiResponse{Message: "Genre does not exist", Data: nil}, http.StatusNotFound, nil
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute query",
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}

	queryAddFavoriteBook := "INSERT INTO favorite_books (user_id, book_id) VALUES (?, ?)"
	stmt, err = repository.db.PrepareContext(ctx, queryAddFavoriteBook)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, userID, bookID)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute query",
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to get rows affected",
		})

		return helpers.ApiResponse{Message: "Failed to get rows affected", Data: nil}, http.StatusInternalServerError, err
	}

	repository.logLogrus.Infof("Success Added Favorite Book Repository. rows affected: %d", rowsAffected)

	favoriteBook = identity.FavoriteBook{
		User: user,
		Book: book,
	}

	return helpers.ApiResponse{Message: "Success added favorite book", Data: favoriteBook}, http.StatusOK, nil
}

func (repository *ManagementBookRepository) GetFavoriteBooksRepository(userID int) (helpers.ApiResponse, int, error) {
	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	favoriteBooksRedis, err := repository.rdb.Get(ctx, fmt.Sprintf("favorite_books: %d", userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			responseRepo, code, err := repository.GetFavoriteBooksDBMysql(userID)
			return responseRepo, code, err
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to get data in redis",
		}).Error("Failed to get data in redis")

		return helpers.ApiResponse{Message: "Failed to get data in redis", Data: nil}, http.StatusInternalServerError, err
	}

	var favoriteBooks []identity.Book
	err = json.Unmarshal([]byte(favoriteBooksRedis), &favoriteBooks)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to unmarshal data redis",
		}).Error("Failed to unmarshal data redis")

		return helpers.ApiResponse{Message: "Failed to unmarshal data redis", Data: nil}, http.StatusInternalServerError, err
	}

	return helpers.ApiResponse{Message: "Success get data in redis", Data: favoriteBooks}, http.StatusOK, nil
}

func (repository *ManagementBookRepository) GetFavoriteBooksDBMysql(userID int) (helpers.ApiResponse, int, error) {
	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	queryFavorite := "SELECT book_id FROM favorite_books WHERE user_id = ?"
	stmt, err := repository.db.PrepareContext(ctx, queryFavorite)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute query",
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}
	defer rows.Close()

	var books []identity.Book

	for rows.Next() {
		var book identity.Book
		err = rows.Scan(&book.BookID)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to scan rows",
			}).Error("Failed to scan rows")

			return helpers.ApiResponse{Message: "Failed to scan rows", Data: nil}, http.StatusInternalServerError, err
		}

		queryBooks := "SELECT title, author, cover, genre_id, isbn, publication_year, quantity FROM books WHERE book_id = ?"
		err = repository.db.QueryRowContext(ctx, queryBooks, book.BookID).Scan(&book.Title, &book.Author, &book.Cover, &book.GenreID, &book.Isbn, &book.PublicationYear, &book.Quantity)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				repository.logLogrus.WithFields(logrus.Fields{
					"error":   err,
					"message": "Book does not exist",
				}).Error("Book does not exist")

				return helpers.ApiResponse{Message: "Book does not exist", Data: nil}, http.StatusNotFound, nil
			}

			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to execute query",
			}).Error("Failed to execute query")

			return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
		}

		queryGenre := "SELECT genre_name FROM genres WHERE genre_id = ?"
		err = repository.db.QueryRowContext(ctx, queryGenre, book.GenreID).Scan(&book.Genre)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				repository.logLogrus.WithFields(logrus.Fields{
					"error":   err,
					"message": "Genre does not exist",
				}).Error("Genre does not exist")

				return helpers.ApiResponse{Message: "Genre does not exist", Data: nil}, http.StatusNotFound, nil
			}

			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to execute query",
			}).Error("Failed to execute query")

			return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
		}

		books = append(books, book)
	}

	booksJson, err := json.Marshal(books)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to marshal books",
		}).Error("Failed to marshal books")

		return helpers.ApiResponse{Message: "Failed to marshal books", Data: nil}, http.StatusInternalServerError, err
	}

	err = repository.rdb.Set(ctx, fmt.Sprintf("favorite_books: %d", userID), booksJson, 30*time.Second).Err()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to set favorite books",
		}).Error("Failed to set favorite books")

		return helpers.ApiResponse{Message: "Failed to set favorite books", Data: nil}, http.StatusInternalServerError, err
	}

	return helpers.ApiResponse{Message: "Success get data favorite books", Data: books}, http.StatusOK, nil
}

func (repository *ManagementBookRepository) DeleteFavoriteBookRepository(userID, bookID int) (helpers.ApiResponse, int, error) {
	var bookName string

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	queryBook := "SELECT title FROM books WHERE book_id = ?"
	stmt, err := repository.db.PrepareContext(ctx, queryBook)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, bookID).Scan(&bookName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Book does not exist",
			}).Error("Book does not exist")

			return helpers.ApiResponse{Message: "Book does not exist", Data: nil}, http.StatusNotFound, nil
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to execute query",
		}).Error("Failed to execute query")

		return helpers.ApiResponse{Message: "Failed to execute query", Data: nil}, http.StatusInternalServerError, err
	}

	query := "DELETE FROM favorite_books WHERE book_id = ? AND user_id = ?"
	stmt, err = repository.db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement", Data: nil}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, bookID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Favorite book does not exist",
			}).Error("Favorite book does not exist")

			return helpers.ApiResponse{Message: "Favorite book does not exist", Data: nil}, http.StatusNotFound, nil
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to delete favorite books",
		}).Error("Failed to delete favorite books")

		return helpers.ApiResponse{Message: "Failed to delete favorite books", Data: nil}, http.StatusInternalServerError, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to get rows affected",
		}).Error("Failed to get rows affected")

		return helpers.ApiResponse{Message: "Failed to get rows affected", Data: nil}, http.StatusInternalServerError, err
	}

	if rowsAffected == 0 {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Rows affected zero",
		}).Error("Rows affected zero")

		return helpers.ApiResponse{Message: "Rows affected zero", Data: nil}, http.StatusInternalServerError, nil
	}

	repository.logLogrus.Infof("Success delected, rows affected: %d", rowsAffected)

	var data = struct {
		BookName string `json:"book_name"`
	}{
		BookName: bookName,
	}

	return helpers.ApiResponse{Message: "Success delete favorite books", Data: data}, http.StatusOK, nil
}
