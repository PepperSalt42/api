package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/gorilla/schema"
	"github.com/jinzhu/gorm"
)

// User contains all informations about an User
type User struct {
	gorm.Model
	SlackID   string `sql:"unique"`
	FirstName string
	LastName  string
	ImageURL  string
	Points    uint
}

// SlackUser contains the data of slack command request
type SlackUser struct {
	ID      string       `json:"id"`
	Profile SlackProfile `json:"profile"`
}

// SlackProfile contains the data contained in SlackUser structure
type SlackProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ImageURL  string `json:"image_192"`
}

// AddUserRequest contains the data of add user request.
type AddUserRequest struct {
	FirstName string `schema:"first_name"`
	LastName  string `schema:"last_name"`
}

// GetUsersTopRequest contains the data of add user request.
type GetUsersTopRequest struct {
	Count int `schema:"count"`
}

func getUser(w http.ResponseWriter, r *http.Request, params martini.Params) {
	id, err := strconv.ParseUint(params["user_id"], 10, 64)
	if err != nil {
		renderJSON(w, http.StatusBadRequest, errInvalidUserID)
		return
	}
	user := &User{}
	if db.First(user, id).Error != nil {
		renderJSON(w, http.StatusNotFound, errUserNotFound)
		return
	}
	renderJSON(w, http.StatusOK, user)
}

func getUsersTop(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := GetUsersTopRequest{
		Count: 6,
	}
	if err := schema.NewDecoder().Decode(&req, q); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	var users []User
	if err := db.Order("points desc").Limit(req.Count).Find(&users).Error; err != nil {
		renderJSON(w, http.StatusNotFound, Error{err.Error()})
		return
	}
	renderJSON(w, http.StatusOK, users)
}

// GetUserBySlackID return an user found in DB using SlackID
// if user doesn't exist yet, we create it
func GetUserBySlackID(id string) (*User, error) {
	user, err := getUserFromSlack(id)
	if err != nil {
		return nil, err
	}
	if err := InsertOrUpdateDB(&User{SlackID: id}, user); err != nil {
		return nil, err
	}
	return user, nil
}

func getUserFromSlack(id string) (*User, error) {
	reqURL := fmt.Sprintf("%s/api/users.info?token=%s&user=%s", slackURL, slackAPIToken, id)
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respData := struct {
		OK    bool      `json:"ok"`
		Error string    `json:"error,omitempty"`
		User  SlackUser `json:"user,omitempty"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, err
	}
	if !respData.OK {
		return nil, errors.New(respData.Error)
	}
	return &User{
		SlackID:   respData.User.ID,
		FirstName: respData.User.Profile.FirstName,
		LastName:  respData.User.Profile.LastName,
		ImageURL:  respData.User.Profile.ImageURL,
		Points:    0,
	}, nil
}
