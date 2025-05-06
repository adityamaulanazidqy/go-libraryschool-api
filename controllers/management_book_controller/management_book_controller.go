package management_book_controller

import (
	"database/sql"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/helpers"
	"go-libraryschool/models/identity"
	"go-libraryschool/models/request_models"
	"go-libraryschool/repository/management_book_repository"
	"net/http"
	"strconv"
)

type ManagementBookController struct {
	db        *sql.DB
	logLogrus *logrus.Logger
	bookRepo  *management_book_repository.ManagementBookRepository
	rdb       *redis.Client
}

func NewManagementBookController(db *sql.DB, logLogrus *logrus.Logger, rdb *redis.Client) *ManagementBookController {
	return &ManagementBookController{db: db, logLogrus: logLogrus, rdb: rdb, bookRepo: management_book_repository.NewManagementBookRepository(db, logLogrus, rdb)}
}

func (controller *ManagementBookController) BookEntityRepository() *management_book_repository.ManagementBookRepository {
	return controller.bookRepo
}

func (controller *ManagementBookController) AddBook(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to parse multipart form",
		}).Error("Failed to parse multipart form")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse multipart form",
		})
		return
	}

	file, handler, err := r.FormFile("cover")
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to get cover image",
		}).Error("Failed to get cover image")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to get cover image",
		})
		return
	}
	defer file.Close()

	filename, err := helpers.SaveImages().Cover(file, handler, "_")
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to save cover image",
		}).Error("Failed to save cover image")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to save cover image",
		})
		return
	}

	genreID, err := strconv.Atoi(r.FormValue("genre_id"))
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to parse genre_id",
		}).Error("Failed to parse genre_id")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse genre_id",
		})
		return
	}

	publicationYear, err := strconv.Atoi(r.FormValue("publication_year"))
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to parse publication_year",
		}).Error("Failed to parse publication_year")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse publication_year",
		})
		return
	}

	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to parse quantity",
		}).Error("Failed to parse quantity")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse quantity",
		})
		return
	}

	book := identity.Book{
		Title:           r.FormValue("title"),
		Author:          r.FormValue("author"),
		Cover:           filename,
		GenreID:         genreID,
		Isbn:            r.FormValue("isbn"),
		PublicationYear: publicationYear,
		Quantity:        quantity,
	}

	responseRepo, code, err := controller.bookRepo.AddBookRepository(r, &book)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, http.StatusOK, responseRepo)
}

// GetBooks godoc
// @Summary GetBooks
// @Description Getting data list books. JWT token is required if you want to use it
// @Tags Books
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /book/get-books [get]
func (controller *ManagementBookController) GetBooks(w http.ResponseWriter, r *http.Request) {
	responseRepo, code, err := controller.BookEntityRepository().GetBooksRepository()
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to get books")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to get books",
			Data:    nil,
		})
		return
	}

	helpers.SendJson(w, code, responseRepo)
}

// GetBook godoc
// @Summary GetBook
// @Description Getting data book. JWT token is required if you want to use it
// @Tags Books
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helpers.ApiResponse
// @Success 400 {object} helpers.ApiResponse
// @Success 404 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /book/get-books [get]
func (controller *ManagementBookController) GetBook(w http.ResponseWriter, r *http.Request) {
	var book request_models.BookById

	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to parse body")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse body",
			Data:    nil,
		})
		return
	}

	responseRepo, code, err := controller.BookEntityRepository().GetBookRepository(r, book.BookID)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}

// DeleteBook godoc
// @Summary Delete Book
// @Description select one of the data books to delete
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request_models.BookById true "bookID"
// @Success 200 {object} helpers.ApiResponse
// @Failure 400 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /book/delete-book [delete]
func (controller *ManagementBookController) DeleteBook(w http.ResponseWriter, r *http.Request) {
	var book request_models.BookById

	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to decode body")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to decode body",
			Data:    nil,
		})
		return
	}

	responseRepo, code, err := controller.BookEntityRepository().DeleteBookRepository(r, book.BookID)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error(responseRepo.Message)

		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}

// UpdateBook godoc
// @Summary Update data Book
// @Description select one of the data books to update
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request_models.BookUpdate false "data update book"
// @Success 200 {object} helpers.ApiResponse
// @Failure 400 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /book/delete-book [delete]
func (controller *ManagementBookController) UpdateBook(w http.ResponseWriter, r *http.Request) {
	var book request_models.BookUpdate

	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to parse body")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to decode body",
			Data:    nil,
		})
		return
	}

	responseRepo, code, err := controller.BookEntityRepository().UpdateBookRepository(r, book)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}

// SearchBooks godoc
// @Summary Searching Books
// @Description Searching books by title book. JWT token is required if you want to use it
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request_models.BookByTitle true "Title Book"
// @Success 200 {object} helpers.ApiResponse
// @Success 400 {object} helpers.ApiResponse
// @Success 404 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /book/search-book [get]
func (controller *ManagementBookController) SearchBooks(w http.ResponseWriter, r *http.Request) {
	var book request_models.BookByTitle
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to parse body")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse body",
			Data:    nil,
		})
		return
	}

	responseRepo, code, err := controller.BookEntityRepository().SearchBooksRepository(r, book.Title)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}

// BorrowedBook godoc
// @Summary Borrowed Books
// @Description Borrowed Book. JWT token is required if you want to use it
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request_models.BookBorrowedRequest true "Data Borrowed"
// @Success 200 {object} helpers.ApiResponse
// @Success 400 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /book/borrowed-book [post]
func (controller *ManagementBookController) BorrowedBook(w http.ResponseWriter, r *http.Request) {
	var book request_models.BookBorrowedRequest

	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error": err,
			"data":  r.Body,
		}).Error("Failed to parse body")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse body",
			Data:    nil,
		})
		return
	}

	responseRepo, code, err := controller.BookEntityRepository().BorrowedBookRepository(r, book)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}

// BookBorrowingData godoc
// @Summary Book Borrowing Data
// @Description Getting data borrowing books. JWT token is required if you want to use it
// @Tags Books
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /book/book-borrowing-data [get]
func (controller *ManagementBookController) BookBorrowingData(w http.ResponseWriter, r *http.Request) {
	responseRepo, code, err := controller.BookEntityRepository().BookBorrowingDataRepository()
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}

// GetBooksCategory godoc
// @Summary Get Books Category
// @Description Getting Books By CategoryID. JWT token is required if you want to use it
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request_models.BookByGenreId true "Category ID"
// @Success 200 {object} helpers.ApiResponse
// @Success 400 {object} helpers.ApiResponse
// @Success 404 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /book/category-books [get]
func (controller *ManagementBookController) GetBooksCategory(w http.ResponseWriter, r *http.Request) {
	var books request_models.BookByGenreId

	err := json.NewDecoder(r.Body).Decode(&books)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to parse body",
		}).Error("Failed to parse body")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse body",
		})
		return
	}

	responseRepo, code, err := controller.BookEntityRepository().GetBooksGenreRepository(books.GenreID)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}
