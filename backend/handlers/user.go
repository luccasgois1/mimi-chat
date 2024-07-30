package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/luccasgois1/mimi-chat/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func isUserInDB(db *gorm.DB, user *models.User) bool{
	var validateUser models.User
	err := db.Where("username = ?", user.Username).First(&validateUser).Error
	return err == nil
}

func isUserValid(user *models.User) bool {
	isUsernameNotEmpty := user.Username != ""
	isPasswordNotEmpty := user.Password != ""
	return isUsernameNotEmpty && isPasswordNotEmpty
}

func loadUserFromRequestBody(w http.ResponseWriter, r *http.Request, user *models.User) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(user)
	defer r.Body.Close()

	handleError(w, err, "Invalid request payload", http.StatusBadRequest)
	return err
}

func setUserPassword(w http.ResponseWriter, user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	handleError(w, err, "Error setting the password", http.StatusInternalServerError)
	if err == nil {
		user.Password = string(hashedPassword)
	}
	return err
}

func createUserInDB(w http.ResponseWriter, db *gorm.DB, user *models.User) error {
	err := db.Create(user).Error
	handleError(w, err, "Error creating user", http.StatusInternalServerError)
	return err
}

func handleError(w http.ResponseWriter, err error, msg string, code int) {
	if err != nil {
		http.Error(w, msg, code)
		log.Println(msg, err)
	}
}

func RegisterHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		if err := loadUserFromRequestBody(w, r, &user); err != nil {
			return
		}
		
		if !isUserValid(&user) {
			handleError(w, models.ErrorInvalidCredentials, "Username or Password are missing.", http.StatusBadRequest)
			return
		}

		if isUserInDB(db, &user) {
			handleError(w, models.ErrorDuplicateUsername, "Username already registed.", http.StatusConflict)
			return
		}

		if err := setUserPassword(w, &user); err != nil {
			return
		}

		if err := createUserInDB(w, db, &user); err != nil {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

func LoginHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestedUser models.User
		var user models.User

		if err := loadUserFromRequestBody(w, r, &requestedUser); err != nil {
			return
		}

		// Check if User exists on DB
		if err := db.Where("username = ?", requestedUser.Username).First(&user).Error; err != nil {
			handleError(w, err, "User not found", http.StatusNotFound)
			return
		}

		// Check if provided password matches with the one in the database
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(requestedUser.Password)); err != nil {
			handleError(w, err, "Invalid password", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}
