package main

type BookService struct {
	repo *BookRepository
}

func NewBookService(r *BookRepository) *BookService {
	return &BookService{r}
}

func (s BookService) ListBooks(f ListBooksFilter) ([]Book, error) {
	return s.repo.ListBooks(f)
}