package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
