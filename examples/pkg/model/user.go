package model

import (
	"gorm.io/gorm"
	"fmt"
)


type User struct {
	gorm.Model
	action.Performer `gorm:"-"`

	Name         string
	Email        string `gorm:"unique;not null"`
}

type EmailDuplicateError struct {
	Email string
}

func (e *EmailDuplicateError) Error() string {
	return fmt.Sprintf("Email '%s' already exists", e.Email)
}
