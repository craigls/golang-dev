package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
)

const (
    minYearPublished = 1500
    maxYearPublished = 2025
    minBookPages     = 10
    maxBookPages     = 2000
)
type BookRepository struct {
	db *sql.DB
}


func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db}
}

type ListBooksFilter struct {
	Authors []int
	Genres []int
	MinPages int
	MaxPages int
    MinYear int
	MaxYear int
    Limit int

}

func (r BookRepository) ListBooks(f ListBooksFilter) ([]Book, error) {

	sql := fmt.Sprintf(`
        SELECT * FROM (
            SELECT
                book.id,
                book.title,
                book.year_published,
                book.rating,
                book.pages,
                author.id AS author_id,
                author.first_name AS author_first_name,
                author.last_name AS author_last_name,
                genre.id AS genre_id,
                genre.title AS genre_title,
                era.id AS era_id,
                era.title AS era_title,
                size.id AS size_id,
                size.title AS size_title,
                ROW_NUMBER() OVER (PARTITION BY book.id ORDER BY book.rating DESC) AS n
            FROM book
            INNER JOIN author ON
                book.author_id = author.id
            INNER JOIN genre ON
                book.genre_id = genre.id
            LEFT JOIN era ON
                book.year_published BETWEEN COALESCE(era.min_year, %d) AND COALESCE(era.max_year, %d)
            LEFT JOIN size ON
                book.pages BETWEEN COALESCE(size.min_pages, %d) AND COALESCE(size.max_pages, %d)
           WHERE 
                (COALESCE(ARRAY_LENGTH($1::int[], 1), 0) = 0 OR author.id = ANY($1::int[]))
                AND (COALESCE(ARRAY_LENGTH($2::int[], 1), 0) = 0 OR genre.id = ANY($2::int[]))
                AND ($3 = 0 OR book.pages >= $3)
                AND ($4 = 0 OR book.pages <= $4)
                AND ($5 = 0 OR book.year_published >= $5)
                AND ($6 = 0 OR book.year_published <= $6)
            ORDER BY
                book.rating DESC
        )
        WHERE
            n = 1
        ORDER BY
            rating DESC
    `, minYearPublished, maxYearPublished, minBookPages, maxBookPages)
    if f.Limit > 0 {
        sql += fmt.Sprintf(" LIMIT %d", f.Limit)
    }

    rows, err := r.db.Query(sql, pq.Array(f.Authors), pq.Array(f.Genres), f.MinPages, f.MaxPages, f.MinYear, f.MaxYear)

    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    books := make([]Book, 0)

    for rows.Next() {
        var (
            book Book
            author Author
            genre Genre
            era Era
            size Size
            n int
        )
        
        err := rows.Scan(
            &book.ID, 
            &book.Title,
            &book.YearPublished,
            &book.Rating,
            &book.Pages,
            &author.ID,
            &author.FirstName,
            &author.LastName,
            &genre.ID,
            &genre.Title,
            &era.ID,
            &era.Title,
            &size.ID,
            &size.Title,
            &n,
        ) 
        book.Author = author
        book.Genre = genre
        
        if err != nil {
            log.Fatal(err)
        }

        books = append(books, book)
    }

    if err := rows.Err(); err != nil {
        // check for errors from iterating rows
        return nil, err
    }

    return books, err      


}