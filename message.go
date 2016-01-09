package main

import (
	"net/http"

	"github.com/gorilla/schema"
	"github.com/jinzhu/gorm"
)

var (
	errInvalidPayload = Error{"Payload is not a valid json"}
	errInvalidCount   = Error{"Invalid count"}
	errInvalidToken   = Error{"Invalid token"}
)

// Message contains all informations about a Message
type Message struct {
	gorm.Model
	UserID  uint
	Message string
}

// SlackMessageRequest is an object version of json received from a Slack request
type SlackMessageRequest struct {
	Token  string `schema:"token"`
	UserID string `schema:"user_id"`
	Text   string `schema:"text"`
}

// addMessage is a route that slack calls to send us new message
func addMessage(w http.ResponseWriter, r *http.Request) {
	var slackMessage SlackMessageRequest
	if err := r.ParseForm(); err != nil {
		renderJSON(w, http.StatusBadRequest, errInvalidPayload)
		return
	}
	if err := schema.NewDecoder().Decode(&slackMessage, r.PostForm); err != nil {
		renderJSON(w, http.StatusBadRequest, errInvalidPayload)
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
