package model

import (
	"fmt"

	"github.com/bioform/ba"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ba.Performer `gorm:"-"`

	Name  string
	Email string `gorm:"unique;not null"`
}

type EmailDuplicateError struct {
	Email string
}

func (e *EmailDuplicateError) Error() string {
	return fmt.Sprintf("Email '%s' already exists", e.Email)
}
