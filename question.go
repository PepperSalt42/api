package main

import (
	"math/rand"
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
	err := db.Order("started_at").Last(question).Error
	return question, err
}

func nextQuestion() error {
	if err := updateUsersPoints(); err != nil {
		return err
	}
	if err := randomizeQuestion(); err != nil {
		return err
	}
	return nil
}

func updateUsersPoints() error {
	q, err := GetCurrentQuestion()
	if err != nil {
		return nil
	}
	tx := db.Begin()
	err = tx.Exec("UPDATE `users` JOIN `answer_entries` ON users.id = answer_entries.user_id SET users.points = users.points + 1 WHERE answer_entries.question_id = ? AND answer_entries.answer_id = ?", q.ID, q.RightAnswerID).Error
	tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func randomizeQuestion() error {
	var questions []Question
	tx := db.Begin()
	defer tx.Commit()
	if err := tx.Where("started_at = ?", 0).Find(&questions).Error; err != nil {
		return err
	}
	question := &questions[rand.Intn(len(questions))]
	return tx.Model(question).UpdateColumn("started_at", time.Now()).Error
}
