package main

import (
	"awesomeProject/database"
	"awesomeProject/handlers"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting database connection...")
	database.ConnectDB() // Подключаемся к базе данных
	log.Println("Database connected successfully.")

	http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for %s", r.Method, r.URL.Path)
		handlers.BooksHandler(w, r)
	}) // Обработчик для работы с книгами

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
