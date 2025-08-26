package main

type Book struct {
	ID int
	Title string
	YearPublished int
	Rating float64
	Pages int
	Genre Genre	
	Author Author
}

type Author struct {
	ID int
	FirstName string
	LastName string
}

type Genre struct {
	ID int
	Title string
}

type Era struct {
	ID int
	Title string
	MinYear int
	MaxYear int
}

type Size struct {
	ID int
	Title string
	MinPages int
	MaxPages int

}