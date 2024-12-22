package handlers

import (
	"awesomeProject/database"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

type Book struct {
	ID     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title  string             `json:"title" bson:"title"`
	Author string             `json:"author" bson:"author"`
}

// Helper function for JSON responses
func jsonResponse(w http.ResponseWriter, status, message string, code int, data ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	// Prepare the response map
	response := map[string]interface{}{
		"status":  status,
		"message": message,
	}

	// If data is provided (non-empty), include it in the response
	if len(data) > 0 {
		response["data"] = data[0]
	}

	// Encode the response to JSON and send it
	json.NewEncoder(w).Encode(response)
}

func BooksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get all books or a specific book by ID
		idParam, ok := r.URL.Query()["id"] // Check if "id" parameter is present
		if ok && idParam[0] == "" {        // If "id" is present but empty
			jsonResponse(w, "fail", "The 'id' parameter cannot be empty", http.StatusBadRequest)
			return
		}
		if ok { // If "id" parameter exists, call getBookByID
			getBookByID(w, r)
		} else { // Otherwise, retrieve all books
			getBooks(w)
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

func createBook(w http.ResponseWriter, r *http.Request) {
	var book Book

	// Декодируем JSON
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		jsonResponse(w, "fail", "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Проверяем обязательные поля
	if book.Title == "" || book.Author == "" {
		jsonResponse(w, "fail", "Missing required fields: title or author", http.StatusBadRequest)
		return
	}

	// Присваиваем уникальный ID
	book.ID = primitive.NewObjectID()

	// Контекст для базы данных
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Вставляем в коллекцию
	_, err := database.BookCollection.InsertOne(ctx, book)
	if err != nil {
		jsonResponse(w, "fail", "Failed to create book", http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	jsonResponse(w, "success", "Successfully created book", http.StatusOK)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	// Get the book ID from the URL query parameter
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		jsonResponse(w, "fail", "Missing ID parameter", http.StatusBadRequest)
		return
	}

	// Convert the ID from hex string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		jsonResponse(w, "fail", "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Decode the book details from the request body
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		jsonResponse(w, "fail", "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set up the filter for the update query
	filter := bson.M{"_id": objectID}

	// Set the update parameters
	update := bson.M{"$set": bson.M{"title": book.Title, "author": book.Author}}

	// Perform the update operation
	result, err := database.BookCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		jsonResponse(w, "fail", "Failed to update book", http.StatusInternalServerError)
		return
	}

	// Check if any document was updated
	if result.MatchedCount == 0 {
		jsonResponse(w, "fail", "Book not found", http.StatusNotFound)
		return
	}

	// Return success response
	jsonResponse(w, "success", "Successfully updated book", http.StatusOK)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	// Get the book ID from the URL query parameter
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		jsonResponse(w, "fail", "Missing ID parameter", http.StatusBadRequest)
		return
	}

	// Convert the ID from hex string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		jsonResponse(w, "fail", "Invalid ID format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Perform the delete operation
	result, err := database.BookCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		jsonResponse(w, "fail", "Failed to delete book", http.StatusInternalServerError)
		return
	}

	// Check if any document was deleted
	if result.DeletedCount == 0 {
		jsonResponse(w, "fail", "Book not found", http.StatusNotFound)
		return
	}

	// Return success response
	jsonResponse(w, "success", "Successfully deleted book", http.StatusOK)
}

func getBooks(w http.ResponseWriter) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query to get all books from the database
	cursor, err := database.BookCollection.Find(ctx, bson.M{})
	if err != nil {
		jsonResponse(w, "fail", "Failed to retrieve books", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var books []Book
	if err = cursor.All(ctx, &books); err != nil {
		jsonResponse(w, "fail", "Error decoding books", http.StatusInternalServerError)
		return
	}

	// Check if no books are found
	if len(books) == 0 {
		jsonResponse(w, "fail", "No books in library", http.StatusNotFound)
		return
	}

	// If books are found, return the books data in JSON format
	jsonResponse(w, "success", "Books retrieved successfully", http.StatusOK, books)
}

func getBookByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		jsonResponse(w, "fail", "Missing ID parameter", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		jsonResponse(w, "fail", "Invalid ID format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var book Book
	err = database.BookCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&book)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			jsonResponse(w, "fail", "Book not found", http.StatusNotFound)
		} else {
			jsonResponse(w, "fail", "Failed to retrieve book", http.StatusInternalServerError)
		}
		return
	}

	// Return the book data in JSON format
	jsonResponse(w, "success", "Book retrieved successfully", http.StatusOK, book)
}
