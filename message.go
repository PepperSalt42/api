package main

import (
	"net/http"

	"github.com/gorilla/schema"
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
	var slackMessage SlackMessageRequest
	if err := r.ParseForm(); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	if err := schema.NewDecoder().Decode(&slackMessage, r.PostForm); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	if slackMessage.Token != slackOutgoingToken {
		renderJSON(w, http.StatusBadRequest, errInvalidToken)
		return
	}
	user, err := getUpdateUserBySlackID(slackMessage.UserID)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = db.Create(&Message{
		UserID:  user.ID,
		Message: slackMessage.Text,
	}).Error
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := GetMessagesRequest{
		FromID: 0,
		Count:  10,
	}
	if err := schema.NewDecoder().Decode(&req, q); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	messages := []Message{}
	if db.Limit(req.Count).Find(&messages, "id > ?", req.FromID).Error != nil {
		renderJSON(w, http.StatusNotFound, errUserNotFound)
		return
	}
	renderJSON(w, http.StatusOK, messages)
}
