package main

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

// Message contains all informations about a Message
type Message struct {
	gorm.Model
	UserID  uint
	Message string
}

// SlackMessageRequest contains the data of slack command request
type SlackMessageRequest struct {
	Token  string `schema:"token"`
	UserID string `schema:"user_id"`
	Text   string `schema:"text"`
}

// GetMessagesRequest contains the data of slack command request
type GetMessagesRequest struct {
	FromID int `schema:"from_id,omitempty"`
	Count  int `schema:"count,omitempty"`
}

// addMessage is a route that slack calls to send us new message
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
	user, err := getUpdateUserBySlackID(req.UserID)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = db.Create(&Message{
		UserID:  user.ID,
		Message: req.Text,
	}).Error
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

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
