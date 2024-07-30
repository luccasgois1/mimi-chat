package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luccasgois1/mimi-chat/handlers"
	"github.com/luccasgois1/mimi-chat/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ServerConfiguration struct {
	URL  string
	PORT uint
}

type DBConfiguration struct {
	FILE string
}

var SERVER_CONFIG ServerConfiguration = ServerConfiguration{"localhost", 8080}
var DB_CONFIG DBConfiguration = DBConfiguration{"mimi-chat.db"}

func main() {

	// Connect to database
	db, err := gorm.Open(sqlite.Open(DB_CONFIG.FILE), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database: ", err)
	}

	// Auto-migrate the User model
	db.AutoMigrate(&models.User{})

	// Set up the router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/api/v1/register", handlers.RegisterHandler(db)).Methods("POST")
	r.HandleFunc("/api/v1/login", handlers.LoginHandler(db)).Methods("POST")

	// Start Server
	portToListen := fmt.Sprintf(":%d", SERVER_CONFIG.PORT)
	fmt.Printf("Server started at http://%s%s\n", SERVER_CONFIG.URL, portToListen)
	log.Fatal(http.ListenAndServe(portToListen, r))
}
