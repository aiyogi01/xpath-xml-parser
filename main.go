package main

import (
	"fmt"
	"io"
	"os"
)

// ---------------------------------------------------------------------------
//	XPath expressions
// ---------------------------------------------------------------------------

// Bookstore contains XPath expressions for extracting book data from an XML file
type Bookstore struct {
	Books []Book `xpath:"bookstore/book"`
}

// Book contains XPath expressions for parsing a book node
type Book struct {
	Title    Title    `xpath:"title"`
	Authors  []string `xpath:"author"`
	Year     string   `xpath:"year"`
	Price    string   `xpath:"price"`
	Metadata Metadata `xpath:"."`
}

// Title contains XPath expressions for parsing the book title
type Title struct {
	Name string `xpath:"text()"`
	Lang string `xpath:"@lang"`
}

// Metadata contains XPath expressoins for parsing additional book data
type Metadata struct {
	Category string `xpath:"@category"`
	Cover    string `xpath:"@cover"`
}

// ---------------------------------------------------------------------------
//	Parse XML file
// ---------------------------------------------------------------------------

// ParseXMLFile parses an XML file and returns a list of books
func ParseXMLFile(r io.Reader) (*Bookstore, error) {
	root, err := ParseXML(r)
	if err != nil {
		return nil, err
	}

	var books Bookstore
	e := Unmarshall(root, &books)
	if e != nil {
		return nil, e
	}
	return &books, nil
}

// ---------------------------------------------------------------------------
//	Main
// ---------------------------------------------------------------------------

func main() {
	file := "tests/books.xml"

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	data, err := ParseXMLFile(f)
	if err != nil {
		panic(err)
	}

	fmt.Println(data)
}
