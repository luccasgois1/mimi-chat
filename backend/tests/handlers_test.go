package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luccasgois1/mimi-chat/handlers"
	"github.com/luccasgois1/mimi-chat/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to test database", err)
	}
	db.AutoMigrate(&models.User{})
	return db
}

func teardownTestDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database object", err)
	}
	if err := sqlDB.Close(); err != nil {
		log.Fatal("failed to close database connection", err)
	}
}

func checkStatusCode(t *testing.T, status int, expectedStatus int) {
	if status != expectedStatus {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatus)
	}
}

func checkResponseBody(t *testing.T, actualResponseBody string, expectedResponseBody string) {
	if actualResponseBody != expectedResponseBody {
		t.Errorf("handler returned wrong status code: got %v want %v", actualResponseBody, expectedResponseBody)
	}
}

func TestRegisterHandler(t *testing.T) {
	// Initial setup
	db := setupTestDB()
	defer teardownTestDB(db)
	handler := handlers.RegisterHandler(db)

	// Close DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database object", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Fatal("failed to close database", err)
		}
	}()

	// Create mock data
	user := models.User{Username: "testuser", Password: "testpass"}
	body, _ := json.Marshal(user)

	// Prepare the register http request
	req, err := http.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Validate if status is the expected one
	checkStatusCode(t, rr.Code, http.StatusCreated)

	// Validate if response body is the created User
	var createdUser models.User
	json.Unmarshal(rr.Body.Bytes(), &createdUser)
	if createdUser.Username != user.Username {
		t.Errorf("Handler returned unexpected body: got %v want %v", createdUser.Username, user.Username)
	}
}

func TestRegisterHandlerInvalidRequestPayload(t *testing.T) {
	// Initial setup
	db := setupTestDB()
	defer teardownTestDB(db)
	handler := handlers.RegisterHandler(db)

	// Create mock data
	body := []byte(`{ "title": "testtext" }`)

	// Prepare the register http request
	req, err := http.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Validate user cannot request with missing data
	checkStatusCode(t, rr.Code, http.StatusBadRequest)
	expectedBody := "Invalid request payload"
	actualBody := strings.TrimSpace(rr.Body.String())
	checkResponseBody(t, actualBody, expectedBody)
}

func TestRegisterHandlerDuplicatedUsername(t *testing.T) {
	// Initial setup
	db := setupTestDB()
	defer teardownTestDB(db)
	handler := handlers.RegisterHandler(db)

	// Create mock data
	body := []byte(`{ "username": "testuser", "password": "testpass"}`)
	bodyFailure := []byte(`{ "username": "testuser", "password": "testpassword"}`)

	// Prepare the register http request
	// First Request to register username
	req, err := http.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Second request to try to register same
	reqFailure, err := http.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(bodyFailure))
	if err != nil {
		t.Fatal(err)
	}
	reqFailure.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, reqFailure)

	// Validate user cannot register twice
	checkStatusCode(t, rr.Code, http.StatusConflict)
	expectedBody := "Username already registed."
	actualBody := strings.TrimSpace(rr.Body.String())
	checkResponseBody(t, actualBody, expectedBody)
}

func TestLoginHandler(t *testing.T) {
	// Initial Setup
	db := setupTestDB()
	defer teardownTestDB(db)
	handler := handlers.LoginHandler(db)

	// Create mock data
	// Add mock username to Database
	user := models.User{Username: "testuser", Password: "testpass"}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("failed to hash the password", err)
	}
	user.Password = string(hashedPassword)
	db.Create(&user)
	// Create Body request
	body := []byte(`{ "username": "testuser", "password": "testpass"}`)

	// Create login request
	req, err := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Validate status code was sucessfull
	checkStatusCode(t, rr.Code, http.StatusOK)
}

func TestLoginHandlerUserDoesNotExists(t *testing.T) {
	// Initial Setup
	db := setupTestDB()
	defer teardownTestDB(db)
	handler := handlers.LoginHandler(db)

	// Create mock data
	// Create Body request
	body := []byte(`{ "username": "testuser", "password": "testpass"}`)

	// Create login request
	req, err := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Validate status code was sucessfull
	checkStatusCode(t, rr.Code, http.StatusNotFound)
	expectedBody := "User not found"
	actualBody := strings.TrimSpace(rr.Body.String())
	checkResponseBody(t, actualBody, expectedBody)
}

func TestLoginHandlerInvalidPassword(t *testing.T) {
	// Initial Setup
	db := setupTestDB()
	defer teardownTestDB(db)
	handler := handlers.LoginHandler(db)

	// Create mock data
	// Add mock username to Database
	user := models.User{Username: "testuser", Password: "testpass"}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("failed to hash the password", err)
	}
	user.Password = string(hashedPassword)
	db.Create(&user)
	// Create Body request
	body := []byte(`{ "username": "testuser", "password": "testpassword"}`)

	// Create login request
	req, err := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Validate status code was sucessfull
	checkStatusCode(t, rr.Code, http.StatusUnauthorized)
	expectedBody := "Invalid password"
	actualBody := strings.TrimSpace(rr.Body.String())
	checkResponseBody(t, actualBody, expectedBody)
}

func TestLoginHandlerInvalidBody(t *testing.T) {
	// Initial Setup
	db := setupTestDB()
	defer teardownTestDB(db)
	handler := handlers.LoginHandler(db)

	// Create mock data
	// Add mock username to Database
	user := models.User{Username: "testuser", Password: "testpass"}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("failed to hash the password", err)
	}
	user.Password = string(hashedPassword)
	db.Create(&user)
	// Create Body request
	body := []byte(`{ "username": "testuser", "password": "testpassword", "adminAccess":true}`)

	// Create login request
	req, err := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Validate status code was sucessfull
	checkStatusCode(t, rr.Code, http.StatusBadRequest)
	expectedBody := "Invalid request payload"
	actualBody := strings.TrimSpace(rr.Body.String())
	checkResponseBody(t, actualBody, expectedBody)
}
