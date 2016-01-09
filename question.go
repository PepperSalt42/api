package main

import (
	"net/http"
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

// getCurrentQuestion returns current question
func getCurrentQuestion(w http.ResponseWriter, r *http.Request) {
	question, err := GetCurrentQuestion()
	if err != nil {
		renderJSON(w, http.StatusNotFound, errCurQuestionNotFound)
	}
	renderJSON(w, http.StatusOK, question)
}

// GetCurrentQuestion returns the current question
func GetCurrentQuestion() (*Question, error) {
	question := &Question{}
	err := db.Order("started_at desc").First(question).Error
	return question, err
}
