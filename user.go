package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
)

// User contains information about an User.
type User struct {
	gorm.Model
	SlackID   string `sql:"unique"`
	FirstName string
	LastName  string
	ImageURL  string
	Points    uint
}

// SlackUser contains the data of slack command request.
type SlackUser struct {
	ID      string       `json:"id"`
	Profile SlackProfile `json:"profile"`
}

// SlackProfile contains the data contained in SlackUser structure.
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

// GetUsersTopRequest contains the data of get users top request.
type GetUsersTopRequest struct {
	Count int `schema:"count"`
}

// getUser returns an user using http protocol.
func getUser(w http.ResponseWriter, r *http.Request, params martini.Params) {
	id, err := strconv.ParseUint(params["user_id"], 10, 64)
	if err != nil {
		renderJSON(w, http.StatusBadRequest, errInvalidUserID)
		return
	}
	user, err := GetUser(uint(id))
	if err != nil {
		renderJSON(w, http.StatusNotFound, errUserNotFound)
		return
	}
	renderJSON(w, http.StatusOK, user)
}

// getUsersTop returns the top list of players (by points).
func getUsersTop(w http.ResponseWriter, r *http.Request) {
	req := GetUsersTopRequest{
		Count: 6,
	}
	if err := decodeRequestQuery(r, &req); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	users, err := GetUsersTop(req.Count)
	if err != nil {
		renderJSON(w, http.StatusNotFound, Error{err.Error()})
		return
	}
	renderJSON(w, http.StatusOK, users)
}

// GetUser returns the user associated to the id.
func GetUser(id uint) (*User, error) {
	user := &User{}
	err := db.First(user, id).Error
	return user, err
}

// GetUserBySlackID returns the user associated to the SlackID.
// if user doesn't exist yet, we create it.
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

// GetUsersTop return the top users by points with maximum count users.
func GetUsersTop(count int) (users []User, err error) {
	err = db.Order("points desc").Limit(count).Find(&users).Error
	return
}

// getUserFromSlack calls slackAPI to get user information and returns an user object.
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
