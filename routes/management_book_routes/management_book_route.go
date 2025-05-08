package management_book_routes

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/controllers/management_book_controller"
	"go-libraryschool/helpers"
	"go-libraryschool/middlewares"
	"net/http"
)

func ManagementBookRoute(mux *http.ServeMux, db *sql.DB, logLogrus *logrus.Logger, rdb *redis.Client) {
	controller := management_book_controller.NewManagementBookController(db, logLogrus, rdb)

	registerRoute := func(path string, method string, roles []string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		mux.Handle(path, middlewares.JWTMiddleware(roles...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				helpers.SendJson(w, http.StatusMethodNotAllowed, helpers.ApiResponse{
					Message: "Method Not Allowed",
				})
				return
			}
			handlerFunc(w, r)
		})))
	}

	registerRoute("/book/add-book", http.MethodPost, []string{"Manager", "Librarian"}, controller.AddBook)
	registerRoute("/book/get-books", http.MethodGet, []string{"Manager", "Librarian", "Student"}, controller.GetBooks)
	registerRoute("/book/get-book", http.MethodGet, []string{"Manager", "Librarian", "Student"}, controller.GetBook)
	registerRoute("/book/search-book", http.MethodGet, []string{"Manager", "Librarian", "Student"}, controller.SearchBooks)
	registerRoute("/book/delete-book", http.MethodDelete, []string{"Manager", "Librarian"}, controller.DeleteBook)
	registerRoute("/book/update-book", http.MethodPut, []string{"Manager"}, controller.UpdateBook)
	registerRoute("/book/borrowed-book", http.MethodPost, []string{"Manager", "Student"}, controller.BorrowedBook)
	registerRoute("/book/book-borrowing-data", http.MethodGet, []string{"Manager", "Librarian"}, controller.BookBorrowingData)
	registerRoute("/book/category-books", http.MethodGet, []string{"Manager", "Librarian", "Student"}, controller.GetBooksCategory)
	registerRoute("/book/add-favorite-book", http.MethodPost, []string{"Librarian", "Student"}, controller.AddFavoriteBook)
	registerRoute("/book/delete-favorite-book", http.MethodDelete, []string{"Librarian", "Student"}, controller.DeleteFavoriteBook)
	registerRoute("/book/get-favorite-book", http.MethodGet, []string{"Librarian", "Student"}, controller.GetFavoriteBooks)
}
