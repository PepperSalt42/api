package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

// Message contains information about a message.
type Message struct {
	gorm.Model
	UserID  uint
	Message string
	SentAt  time.Time
}

// SlackMessageRequest contains the data of slack command request.
type SlackMessageRequest struct {
	Token     string `schema:"token"`
	Timestamp string `schema:"timestamp"`
	UserID    string `schema:"user_id"`
	Text      string `schema:"text"`
}

// GetMessagesRequest contains the data of slack command request.
type GetMessagesRequest struct {
	FromID int `schema:"from_id,omitempty"`
	Count  int `schema:"count,omitempty"`
}

// addMessage adds a message in the database.
func addMessage(w http.ResponseWriter, r *http.Request) {
	var req SlackMessageRequest
	if err := decodeRequestForm(r, &req); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	if req.Token != slackOutgoingToken {
		renderJSON(w, http.StatusBadRequest, errInvalidToken)
		return
	}
	timestamp, err := strconv.ParseFloat(req.Timestamp, 32)
	if err != nil {
		renderJSON(w, http.StatusBadRequest, errInvalidTimestamp)
	}
	user, err := GetUserBySlackID(req.UserID)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = db.Create(&Message{
		UserID:  user.ID,
		Message: req.Text,
		SentAt:  time.Unix(int64(timestamp), 0),
	}).Error
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// getMessages returns the messages contained in the database.
func getMessages(w http.ResponseWriter, r *http.Request) {
	req := GetMessagesRequest{
		FromID: 0,
		Count:  10,
	}
	if err := decodeRequestQuery(r, &req); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	var messages []Message
	if db.Order("id desc").Limit(req.Count).Find(&messages, "id > ?", req.FromID).Error != nil {
		renderJSON(w, http.StatusNotFound, errMessagesNotFound)
		return
	}
	renderJSON(w, http.StatusOK, messages)
}
