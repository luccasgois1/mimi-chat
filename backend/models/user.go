package models

import (
	"errors"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
}

// Custom error messages
var (
	ErrorInvalidCredentials = errors.New("invalid credentials")
	ErrorDuplicateUsername = errors.New("username already registed")
)
