package main

import (
	"awesomeProject/database"
	"awesomeProject/handlers"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting database connection...")
	database.ConnectDB()
	log.Println("Database connected successfully.")

	// Обработчик для статических файлов (HTML, CSS, JS)
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	// Обработчик для работы с книгами (API)
	http.HandleFunc("/books", handlers.BooksHandler)

	// Запускаем сервер на порту 8080
	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
