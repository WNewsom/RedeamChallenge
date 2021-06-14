package main

//These some tests will fail after the first time running, as the containeris still active and has updated to the changes.
import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

/*
This test will break if book 1 gets deleted or its title/ isCheckedIn is changed.
*/
func TestRead(t *testing.T) {
	readFromFile("books.csv")
	if Books[0].Title != "Book 1" {
		t.Error("Books title is incorrect. Recieved: ", Books[0].Title, "Expected : Book 1")
	}
	if Books[0].IsCheckedIn != true {
		t.Error("Books isCheckedIn is incorrect. Recieved: ", Books[0].IsCheckedIn, "Expected : Book 1")
	}
}

func TestHomePage(t *testing.T) {

	resp, err := http.Get("http://localhost")

	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Error("Expeced Response code 200. Recieved ", resp.StatusCode)
	}

}

/*
This test makes sure that all URLs not used in the API result in 404
*/
func TestBadPage(t *testing.T) {

	resp, err := http.Get("http://localhost/whateverurl")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Expeced Response code 404. Recieved ", resp.StatusCode)
	}

}

/*
Tests that the /books endpoint returns code 200 and all of the books in the csv file
*/
func TestAllBooks(t *testing.T) {

	resp, err := http.Get("http://localhost/books")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Expeced Response code 200. Recieved ", resp.StatusCode)
	}

	expectedBody := strings.TrimSpace(`[{"title":"Book 1","author":"Author 1","publisher":"publisher","publishdate":"11111111","rating":1,"ischeckedin":true},{"title":"Book 2","author":"Author 2","publisher":"publisher","publishdate":"11111112","rating":3,"ischeckedin":false}]`)

	defer resp.Body.Close()

	body, err1 := io.ReadAll(resp.Body)
	if err1 != nil {
		t.Fatal(err1)
	}
	bodyStr := strings.TrimSpace(string(body))

	if expectedBody != bodyStr {
		t.Error("All of the books were not returned correctly. Recieved \n", bodyStr, "\n wanted \n", expectedBody)
	}

}

/*
All books only accepts GET requests, so making sure a POST command results in error.
*/
func TestBadRequestAllBooks(t *testing.T) {
	resp, err := http.PostForm("http://localhost/books", url.Values{"title": {"Book 1"}, "author": {"Author 1"}})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Error("Expeced Response code", http.StatusMethodNotAllowed, "Recieved ", resp.StatusCode)
	}

}

/*
Adds a new book using a post
*/
func TestNewBookPost(t *testing.T) {
	data := url.Values{}

	data.Set("title", "book 10")
	data.Set("author", "author 1")
	data.Set("publisher", "publisher 4")
	data.Set("publishdate", "12345678")
	data.Set("rating", "3")
	data.Set("ischeckedin", "true")

	resp, err := http.PostForm("http://localhost/new", data)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {

		body, err1 := io.ReadAll(resp.Body)
		if err1 != nil {
			t.Fatal(err1)
		}
		bodyStr := string(body)
		fmt.Println("\n", bodyStr)
		t.Error("Expeced Response code", http.StatusCreated, "Recieved ", resp.StatusCode)
	}

}

/*
Tests when bad data is sent
*/
func TestNewBadData(t *testing.T) {
	data := url.Values{}

	data.Set("title", "book 10")
	data.Set("author", "author 1")
	data.Set("publisher", "publisher 4")
	data.Set("publishdate", "1234567") // needs 8, not 7
	data.Set("rating", "3")
	data.Set("ischeckedin", "true")

	resp, err := http.PostForm("http://localhost/new", data)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Expeced Response code", http.StatusBadRequest, "Recieved ", resp.StatusCode)
	}
}

func TestNewBadData2(t *testing.T) {
	data := url.Values{}

	data.Set("title", "book 10")
	data.Set("author", "author 1")
	data.Set("publisher", "publisher 4")
	data.Set("publishdate", "IShouldBeAnInt") //not an int
	data.Set("rating", "3")
	data.Set("ischeckedin", "true")

	resp, err := http.PostForm("http://localhost/new", data)

	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Expeced Response code", http.StatusBadRequest, "Recieved ", resp.StatusCode)
	}
}

func TestNewBadData3(t *testing.T) {
	data := url.Values{}

	data.Set("title", "book 10")
	data.Set("author", "author 1")
	data.Set("publisher", "publisher 4")
	data.Set("publishdate", "12345678")
	data.Set("rating", "3")
	data.Set("ischeckedin", "IShoudBeABool") //not a bool

	resp, err := http.PostForm("http://localhost/new", data)

	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Expeced Response code", http.StatusBadRequest, "Recieved ", resp.StatusCode)
	}
}

func TestNewBadData4(t *testing.T) {
	data := url.Values{}

	data.Set("title", "") //no title
	data.Set("author", "author 1")
	data.Set("publisher", "publisher 4")
	data.Set("publishdate", "12345678")
	data.Set("rating", "3")
	data.Set("ischeckedin", "true")
	resp, err := http.PostForm("http://localhost/new", data)

	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Expeced Response code", http.StatusBadRequest, "Recieved ", resp.StatusCode)
	}
}

func TestBadNewMethodNotFound(t *testing.T) {
	resp, err := http.Get("http://localhost/new")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Error("Expeced Response code 405. Recieved ", resp.StatusCode)
	}

}

/*
If the book 1 is not present then this test will fail
*/
func TestSingleBookGet(t *testing.T) {
	resp, err := http.Get("http://localhost/books/book-2")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Expeced Response code 200. Recieved ", resp.StatusCode)
	}

	body, err1 := io.ReadAll(resp.Body)
	if err1 != nil {
		t.Fatal(err1)
	}
	bodyStr := strings.TrimSpace(string(body))

	expectedBody := `{"title":"Book 2","author":"Author 2","publisher":"publisher","publishdate":"11111112","rating":3,"ischeckedin":false}`
	if expectedBody != bodyStr {
		t.Error("All of the books were not returned correctly. Recieved \n", bodyStr, "\n", expectedBody, strings.Compare(bodyStr, expectedBody))
	}
}

func TestSingleBookNotFound(t *testing.T) {
	resp, err := http.Get("http://localhost/books/Great-Gatsby")

	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Expeced Response code 404. Recieved ", resp.StatusCode)
	}

}

/*
Now we are testing the patch functionality.
*/
func TestPatch(t *testing.T) {

	data := url.Values{}

	data.Set("author", "newAuthor")

	req, err := http.NewRequest("PATCH", "http://localhost/books/book-1", strings.NewReader(data.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	defer req.Body.Close()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Expeced Response code", http.StatusOK, "Recieved ", resp.StatusCode)
	}
	resp2, err := http.Get("http://localhost/books/book-1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	body, err1 := io.ReadAll(resp2.Body)
	if err1 != nil {
		t.Fatal(err1)
	}
	bodyStr := strings.TrimSpace(string(body))

	expectedBody := strings.TrimSpace(`{"title":"Book 1","author":"newAuthor","publisher":"publisher","publishdate":"11111111","rating":1,"ischeckedin":true}`)
	if bodyStr != expectedBody {
		t.Error("The book was not returned correctly. Recieved \n", bodyStr, "expected\n", expectedBody, strings.Compare(bodyStr, expectedBody))
	}

}

func TestDelete(t *testing.T) {

	req, err := http.NewRequest("DELETE", "http://localhost/books/book-1", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Expeced Response code", http.StatusOK, "Recieved ", resp.StatusCode)
	}
	resp2, err := http.Get("http://localhost/books/book-1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Error("Expeced Response code", http.StatusNotFound, "Recieved ", resp.StatusCode)
	}

}

/*
Test the writing method. Its not used because it can break the other tests.
*/
// func TestWrite(t *testing.T) {

// 	readFromFile("books.csv")

// 	if Books[0].Title != "Book 1" {
// 		t.Error("Books title is incorrect. Recieved: ", Books[0].Title, "Expected : Book 1")
// 	}
// 	if Books[0].IsCheckedIn != true {
// 		t.Error("Books isCheckedIn is incorrect. Recieved: ", Books[0].IsCheckedIn, "Expected : Book 1")
// 	}
// 	var newBook Book
// 	newBook.Title = "This"
// 	newBook.Author = "is"
// 	newBook.Publisher = "aBook"
// 	newBook.PublishDate = "12345678"
// 	newBook.Rating = 3
// 	newBook.IsCheckedIn = true

// 	Books = append(Books, newBook)
// 	writeToFile("books.csv")

// 	readFromFile("books.csv")
// 	if Books[2].Title != "This"{
// 		t.Error("Writing test failed")

// 	}

// }
