package main

import (
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
)

var (
	errUserNotFound  = Error{Error: "User not found"}
	errInvalidUserID = Error{Error: "Invalid user_id"}
)

// User contains all informations about an User
type User struct {
	gorm.Model
	SlackID    uint
	Name       string
	PictureURL string
	Points     uint
}

func addUser(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	user := &User{Name: name}
	db.Create(user)
	renderJSON(w, http.StatusCreated, user)
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
