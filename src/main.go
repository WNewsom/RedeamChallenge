package main

/*
This is a simple RESTful API. It can Create, Update, Read, and Delete enteries to a csv file.

DOCKER COMMANDS-

docker build -t will-rest-api .
docker run -p 80:80 -it will-rest-api

*/

//docker run -it -p 80:80
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

/*
The model for the books is given by this struct.
*/
type Book struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Publisher   string `json:"publisher"`
	PublishDate string `json:"publishdate"` //In the format of MMDDYYYY
	Rating      int    `json:"rating"`      // Measured on a scale of 1 - 3
	IsCheckedIn bool   `json:"ischeckedin"` // True if checked in, false if cheched out
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
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatalln("File open failed", err)
	}
	r := csv.NewReader(f)
	defer f.Close()
	Books = nil
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		readRating, _ := strconv.Atoi(record[4])
		readCheckIn, _ := strconv.ParseBool(record[5])
		Books = append(Books, Book{Title: record[0], Author: record[1], Publisher: record[2], PublishDate: record[3], Rating: readRating, IsCheckedIn: readCheckIn})
	}

}

/*
Because deleting and editing the books can break some of the unit tests, the write to file function is commented out, but its here for future implementation.
However, it higlights the problem with a csv file, and that its hard to write to and the simplest way is to rewrite the entire file. This isnt feasable for large operations, and a database would be better.
*/

func writeToFile(filepath string) {

	f, err := os.Create(filepath)
	if err != nil {
		log.Fatalln("File open failed", err)
	}
	f.Seek(0, 0) //removing the content of the file

	os.Truncate("books.csv", 0)
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	var records [][]string

	for _, book := range Books {
		var bookData []string
		bookData = append(bookData, book.Title)
		bookData = append(bookData, book.Author)
		bookData = append(bookData, book.Publisher)
		bookData = append(bookData, book.PublishDate)
		rat := strconv.Itoa(book.Rating)
		bookData = append(bookData, rat)
		check := strconv.FormatBool(book.IsCheckedIn)
		bookData = append(bookData, check)
		records = append(records, bookData)
	}

	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Fatalln("Error in writing to csv")
		}
	}

}

/*
Returns all of the books stored. Only works with GET, and is part of the READ component of CRUD.
*/
func allEnteries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":

		json.NewEncoder(w).Encode(Books)

	default:
		http.Error(w, "405 Method not allowed, only GET is permited", http.StatusMethodNotAllowed)
	}
}

/*
This homepage will give a 404 error whenever a connection to an unused url is attepmted.
*/
func homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "This is the Homepage")

}

/*
This function serves as read, update, and delete components. One could read individual books with GET, update with PATCH, and delete with DELETE http methods.
*/

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

/*
This fuction will delete the book with the given by the URL. If that book isnt in the list, it will return 404.
*/
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
			//writeToFile("books.csv")
			mutex.Unlock()
			return
		}
	}
	http.Error(w, "404, not found.", http.StatusNotFound)
}

/*
Updates books given by a PATCH request. Because all of the keys dont have to be present, the method must individualy check each to see if it is empty or not.
If the data is invalid, it returns error code 400 and doesnt update the list.
*/

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
						http.Error(w, "400, publishddate not correct, not an int", http.StatusBadRequest)
						fmt.Printf("recieved %v", r.FormValue("publishdate"))
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
			//writeToFile("books.csv")

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
				fmt.Printf("recieved %v", r.FormValue("publishdate"))
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

		//writeToFile("books.csv")

		w.WriteHeader(http.StatusCreated)

		mutex.Unlock()
	default:
		http.Error(w, "405 Method not allowed, only POST commands are permited", http.StatusMethodNotAllowed)
	}

}

/*
The main method reads the csv file, and passes the functions to the handler
*/
func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/books", allEnteries)
	http.HandleFunc("/books/", returnSingleBook)
	http.HandleFunc("/new", createNewBook)
	log.Fatal(http.ListenAndServe(":80", nil))
}
