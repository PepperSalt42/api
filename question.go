package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Question contains all informations about a question
type Question struct {
	gorm.Model
	UserID        uint
	Sentence      string
	RightAnswerID uint
	StartedAt     time.Time
}

// GetCurrentQuestion returns the current question
func GetCurrentQuestion() (*Question, error) {
	question := &Question{}
	err := db.Order("started_at desc").First(question).Error
	return question, err
}
