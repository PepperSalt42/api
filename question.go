package main

import (
	"errors"
	"math/rand"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
)

var errNoQuestionAvailable = errors.New("No question available")

// Question contains information about a question.
type Question struct {
	gorm.Model
	UserID        uint
	Sentence      string
	RightAnswerID uint `json:"-"`
	StartedAt     time.Time
}

// GetCurrentQuestionAnswer contains the data of get current question request.
type GetCurrentQuestionAnswer struct {
	Question *Question
	Answers  []string
}

// getCurrentQuestion returns current question.
func getCurrentQuestion(w http.ResponseWriter, r *http.Request) {
	question, err := GetCurrentQuestion()
	if err != nil {
		renderJSON(w, http.StatusNotFound, errCurQuestionNotFound)
		return
	}
	answers, err := GetAnswersByQuestionID(question.ID)
	if err != nil {
		renderJSON(w, http.StatusNotFound, errCurQuestionNotFound)
		return
	}
	resp := GetCurrentQuestionAnswer{Question: question}
	for _, answer := range answers {
		resp.Answers = append(resp.Answers, answer.Sentence)
	}
	renderJSON(w, http.StatusOK, &resp)
}

// GetCurrentQuestion returns the current question.
func GetCurrentQuestion() (*Question, error) {
	return getCurrentQuestionWithTX(&db)
}

// getCurrentQuestionWithTX returns the current question using database transaction.
func getCurrentQuestionWithTX(tx *gorm.DB) (*Question, error) {
	question := &Question{}
	err := tx.Order("started_at").Last(question).Error
	return question, err
}

// refreshQuestion updates current question every X time.
func refreshQuestion() {
	questionRefreshRate := os.Getenv("QUESTION_REFRESH_RATE")
	wait, err := time.ParseDuration(questionRefreshRate)
	if err != nil {
		log.Fatal("Can't convert question refresh rate")
	}
	for {
		time.Sleep(wait)
		if err := nextQuestion(); err != nil {
			log.WithField("err", err).Error("Can't set nextQuestion")
		}
	}
}

// nextQuestion selects a new random question and updates users points.
func nextQuestion() error {
	tx := db.Begin()
	defer tx.Commit()
	nextQuestion, err := getNextQuestion(tx)
	if err != nil {
		return err
	}
	if err := updateUsersPoints(tx); err != nil {
		return err
	}
	return tx.Model(nextQuestion).UpdateColumn("started_at", time.Now()).Error
}

// updateUsersPoints updates users points.
func updateUsersPoints(tx *gorm.DB) error {
	q, err := getCurrentQuestionWithTX(tx)
	if err != nil {
		return nil
	}
	return tx.Exec("UPDATE `users` JOIN `answer_entries` ON users.id = answer_entries.user_id SET users.points = users.points + 1 WHERE answer_entries.question_id = ? AND answer_entries.answer_id = ?", q.ID, q.RightAnswerID).Error
}

// getNextQuestion returns the next random question.
func getNextQuestion(tx *gorm.DB) (*Question, error) {
	var questions []Question
	if err := tx.Where("started_at = ?", 0).Find(&questions).Error; err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, errNoQuestionAvailable
	}
	return &questions[rand.Intn(len(questions))], nil
}
