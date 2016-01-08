package main

import (
	"github.com/jinzhu/gorm"
)

// Answer contains all informations about an Answer
type Answer struct {
	gorm.Model
	Sentence   string
	QuestionID uint
}
