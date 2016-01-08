package main

import (
	"github.com/jinzhu/gorm"
	"time"
)

// AnswerHistory contains all informations about an AnswerHistory
type AnswerHistory struct {
	gorm.Model
	PlayerID   int
	QuestionID int
	UpdatedAt  time.Time
}
