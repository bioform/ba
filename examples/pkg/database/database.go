package database

import (
	"gorm.io/gorm"
)

var defaultDB *gorm.DB

func Default() *gorm.DB {
	return defaultDB
}
