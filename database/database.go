package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var BookCollection *mongo.Collection // Глобальная переменная для коллекции

func ConnectDB() {
	// Настройка подключения
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Attempting to connect to MongoDB...")
	// Подключение к серверу MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Проверка подключения
	log.Println("Pinging MongoDB server...")
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB server: %v", err)
	}
	log.Println("Successfully connected and pinged MongoDB server.")

	// Привязка коллекции
	BookCollection = client.Database("library").Collection("books")
	log.Println("Book collection initialized.")
}
