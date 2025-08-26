package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	handler := http.NewServeMux()
	booksHandler := NewBooksHandler(db)
	handler.Handle("/api/v1/", http.StripPrefix("/api/v1", booksHandler.CreateRoutes()))
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), handler))

}
