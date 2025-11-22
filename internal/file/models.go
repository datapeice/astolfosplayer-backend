package file

import (
	"gorm.io/gorm"
)

type Track struct {
	gorm.Model
	Hash     string `gorm:"uniqueIndex"`
	Filename string
	Title    string
	Artist   string
	Album    string
	Duration int32
}
