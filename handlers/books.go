package handlers

import (
	"awesomeProject/database"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Book структура для представления книги
type Book struct {
	ID     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title  string             `json:"title" bson:"title"`
	Author string             `json:"author" bson:"author"`
}

// BooksHandler обрабатывает все запросы, связанные с книгами
func BooksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		idParam := r.URL.Query().Get("id")
		if idParam != "" {
			getBookByID(w, r) // Если есть параметр id, вызываем функцию для получения книги по ID
		} else {
			getBooks(w) // Если нет параметра id, выводим все книги
		}
	case http.MethodPost:
		createBook(w, r)
	case http.MethodPut:
		updateBook(w, r)
	case http.MethodDelete:
		deleteBook(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getBooks(w http.ResponseWriter) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.BookCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to retrieve books", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var books []Book
	if err = cursor.All(ctx, &books); err != nil {
		http.Error(w, "Error decoding books", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func getBookByID(w http.ResponseWriter, r *http.Request) {
	// Извлечение ID из URL-параметров
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	// Преобразование строки в ObjectID
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Создание контекста для запроса
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Поиск книги в базе данных по ID
	var book Book
	err = database.BookCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&book)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Book not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve book", http.StatusInternalServerError)
		}
		return
	}

	// Отправка найденной книги в ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	book.ID = primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.BookCollection.InsertOne(ctx, book)
	if err != nil {
		http.Error(w, "Failed to create book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": book.ID}
	update := bson.M{"$set": bson.M{"title": book.Title, "author": book.Author}}

	result, err := database.BookCollection.UpdateOne(ctx, filter, update)
	if err != nil || result.MatchedCount == 0 {
		http.Error(w, "Failed to update book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": book.ID}
	result, err := database.BookCollection.DeleteOne(ctx, filter)
	if err != nil || result.DeletedCount == 0 {
		http.Error(w, "Failed to delete book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
