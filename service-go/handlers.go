package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	service *BookService
}

func NewBooksHandler (db *sql.DB) *Handler {
	repo := NewBookRepository(db)
	service := NewBookService(repo)
	return &Handler{service: service}
}

func (h Handler) CreateRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/books", h.ListBooks)
	return mux
}

func (h Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	var f ListBooksFilter
	var err error

	q := r.URL.Query()
	// TODO: Handle errors
	f.Authors, err = ParseInts(q.Get("authors"))
	f.Genres, err = ParseInts(q.Get("genres"))
	f.MinPages, err = strconv.Atoi(q.Get("min-pages"))
	f.MaxPages, err = strconv.Atoi(q.Get("max-pages"))
	f.MinYear, err = strconv.Atoi(q.Get("min-year"))
	f.MaxYear, err = strconv.Atoi(q.Get("max-year"))
	f.Limit, err = strconv.Atoi(q.Get("limit"))

	books, err := h.service.ListBooks(f)
	
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}