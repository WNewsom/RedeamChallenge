package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Book struct {
	Title       string `json:"Title"`
	Author      string `json:"Author"`
	Publisher   string `json:"publisher"`
	PublishDate string `json:"publishDate"` //In the format of MMDDYYYY
	Rating      int    `json:"rating"`      // Measured on a scale of 1 - 3
	IsCheckedIn bool   `json:"status"`      // True if checked in, false if cheched out
}

var (
	/*
		The mutex is good to have when dealing with delete and edit requests,
		which could create race conditions. Could be improved if there are mutiple
		writers in the system by adding a hierarcical order of semaphores, but its unneccessary in this
		case as were are hosting through localhost.
	*/
	mutex sync.Mutex
	Books []Book //  The list of books read in from our csv "database"
)

func main() {

	readFromFile("books.csv")
	handleRequests()

}

/*
The csv file acts as the database behind the API. This will read the csv and then store the books in a slice.
*/
func readFromFile(filepath string) {
	f, _ := os.Open(filepath)
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(record)
		// fmt.Println(len(record))
		// for value := range record {

		// 	fmt.Printf("  %v\n", record[value])
		// }
		readRating, _ := strconv.Atoi(record[4])
		readCheckIn, _ := strconv.ParseBool(record[5])
		Books = append(Books, Book{Title: record[0], Author: record[1], Publisher: record[2], PublishDate: record[3], Rating: readRating, IsCheckedIn: readCheckIn})
	}

}

func allEnteries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":

		json.NewEncoder(w).Encode(Books)

	default:
		http.Error(w, "405 Method not allowed, only GET is permited", http.StatusMethodNotAllowed)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Welcome to the HomePage!")

}

func returnSingleBook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":

		if r.URL.Path == "/books/" {
			allEnteries(w, r)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/books/")

		for _, book := range Books {
			idTest := strings.ReplaceAll(book.Title, " ", "-")
			fmt.Println(id + " compared to" + idTest)

			if strings.EqualFold(id, idTest) {
				json.NewEncoder(w).Encode(book)
				fmt.Println("test")
				return
			}
		}
		http.Error(w, "404, not found.", http.StatusNotFound)

	case "PATCH":
		patchBook(w, r)

	case "DELETE":
		deleteBook(w, r)
	default:
		http.Error(w, "405 Method not allowed, only PATCH, DELETE, and GET are permited", http.StatusMethodNotAllowed)

	}
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/books/")

	for i, book := range Books {
		idTest := strings.ReplaceAll(book.Title, " ", "-")
		fmt.Println(id + " compared to" + idTest)

		if strings.EqualFold(id, idTest) {
			mutex.Lock()
			//Last element
			if i == len(Books) {
				Books = Books[:len(Books)-1]

				//one of the middle elements or first element
			} else {
				Books = append(Books[:i], Books[i+1:]...)
			}
			fmt.Fprintf(w, "Book: "+idTest+" deleted!")
			//TODO WRITE TO FILE
			mutex.Unlock()
			return
		}
	}
	http.Error(w, "404, not found.", http.StatusNotFound)
}

func patchBook(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/books/")
	for i, book := range Books {
		idTest := strings.ReplaceAll(book.Title, " ", "-")
		fmt.Println(id + " compared to" + idTest)

		if strings.EqualFold(id, idTest) {

			mutex.Lock()
			var newBook = book
			r.ParseForm()
			if r.FormValue("title") != "" { // Form value returns an empty string if the key value isn't present
				newBook.Title = r.FormValue("title")
			}

			if r.FormValue("author") != "" {
				newBook.Author = r.FormValue("author")
			}

			if r.FormValue("publisher") != "" {
				newBook.Publisher = r.FormValue("publisher")
			}

			if r.FormValue("publishdate") != "" {
				if len(r.FormValue("publishdate")) == 8 {
					_, err := strconv.Atoi(r.FormValue("publishdate")) //Dont care about the integer value returned, just making sure that there are only numbers in the publishdate
					if err != nil {
						http.Error(w, "400, publishddate not correct", http.StatusBadRequest)
						mutex.Unlock()
						return
					}
					newBook.PublishDate = r.FormValue("publishdate")
				} else {
					http.Error(w, "400, publishddate not correct", http.StatusBadRequest)
					mutex.Unlock()
					return
				}

			}

			if r.FormValue("rating") != "" {
				rating, err := strconv.Atoi(r.FormValue("rating"))
				if err != nil {
					http.Error(w, "400, rating not correct", http.StatusBadRequest)
					mutex.Unlock()
					return
				}
				if rating == 1 || rating == 2 || rating == 3 {
					newBook.Rating = rating
				} else {
					http.Error(w, "400, rating not on 1-3 scale", http.StatusBadRequest)
					mutex.Unlock()
					return
				}
			}

			if r.FormValue("rating") != "" {
				checkin, err := strconv.ParseBool(r.FormValue("ischeckedin"))

				if err != nil {
					http.Error(w, "400, ischeckedin not a boolean", http.StatusBadRequest)
					mutex.Unlock()
					return
				} else {
					newBook.IsCheckedIn = checkin
				}
			}
			Books[i] = newBook
			json.NewEncoder(w).Encode(Books[i])

			for value := range Books {

				fmt.Printf("  %v\n", Books[value])
			}
			//TODO Update File

			mutex.Unlock()
			return
		}
	}
	http.Error(w, "404, not found.", http.StatusNotFound)
}

/*
This method will take the inputs passed via url encoding to create a new book. If the values are valid, it will and the book to our slice and then write it to our csv "database."
*/
func createNewBook(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	/*
		This method will only use the POST http keyword. If anyother keyword is used, it will result in a 405 error.
	*/
	case "POST":
		mutex.Lock()
		r.ParseForm()
		for key, value := range r.Form {
			fmt.Printf("%s = %s\n", key, value)
		}
		var newBook Book

		/*
			All books must have a title, as we are using that as the key value for finding the books.
		*/
		if r.FormValue("title") != "" {
			newBook.Title = r.FormValue("title")
		} else {
			http.Error(w, "400, All books must have a title", http.StatusBadRequest)
			mutex.Unlock()
			return
		}

		newBook.Author = r.FormValue("author")
		newBook.Publisher = r.FormValue("publisher")

		/*
			Because '-' would conflict in URL encoding, the dates are given in the format of MMDDYYYY.
			These have to be 8 characters long.
		*/
		if len(r.FormValue("publishdate")) == 8 {
			_, err := strconv.Atoi(r.FormValue("publishdate")) //Dont care about the integer value returned, just making sure that there are only numbers in the publishdate
			if err != nil {
				http.Error(w, "400, publishddate not correct", http.StatusBadRequest)
				mutex.Unlock()
				return
			}
			newBook.PublishDate = r.FormValue("publishdate")

		} else {
			http.Error(w, "400, publishddate not correct", http.StatusBadRequest)
			mutex.Unlock()
			return
		}

		/*
			Similar concept with the rating and isChecked in, they have to be able to be parsed as an integer (1-3 in this case) or a boolean respectivly
		*/
		rating, err := strconv.Atoi(r.FormValue("rating"))
		if err != nil {
			http.Error(w, "400, rating not correct", http.StatusBadRequest)
			mutex.Unlock()
			return
		}
		if rating == 1 || rating == 2 || rating == 3 {
			newBook.Rating = rating
		} else {
			http.Error(w, "400, rating not on 1-3 scale", http.StatusBadRequest)
			mutex.Unlock()
			return
		}

		checkin, err := strconv.ParseBool(r.FormValue("ischeckedin"))

		if err != nil {

			http.Error(w, "400, ischeckedin not a boolean", http.StatusBadRequest)
			mutex.Unlock()
			return
		} else {
			newBook.IsCheckedIn = checkin
		}

		/*
			The string has been sucessfuly parsed, and there are no errors with it. We can now add it to the list of books and writed to the csv file.
		*/
		Books = append(Books, newBook)

		// for value := range Books {

		// 	fmt.Printf("  %v\n", Books[value])
		// }

		//TODO Write to csv

		w.WriteHeader(http.StatusCreated)

		mutex.Unlock()
	default:
		http.Error(w, "405 Method not allowed, only POST commands are permited", http.StatusMethodNotAllowed)
	}

}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/books", allEnteries)
	http.HandleFunc("/books/", returnSingleBook)
	http.HandleFunc("/new", createNewBook)
	log.Fatal(http.ListenAndServe(":80", nil))
}
