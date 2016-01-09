package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/schema"
	"github.com/jinzhu/gorm"
)

var (
	db gorm.DB

	slackAPIToken      = os.Getenv("SLACK_API_TOKEN")
	slackOutgoingToken = os.Getenv("SLACK_OUTGOING_TOKEN")
	slackURL           = "https://slack.com"

	errInvalidToken     = Error{"Invalid token"}
	errInvalidUserID    = Error{"Invalid user_id"}
	errMessagesNotFound = Error{"Messages not found"}
	errUserNotFound     = Error{"User not found"}
)

func initDB() {
	var err error
	user := os.Getenv("DB_ENV_MYSQL_USER")
	password := os.Getenv("DB_ENV_MYSQL_PASSWORD")
	dbname := os.Getenv("DB_ENV_MYSQL_DATABASE")
	sqlConnection := fmt.Sprintf("%s:%s@tcp(db:3306)/%s?charset=utf8&parseTime=True&loc=Local", user, password, dbname)
	db, err = gorm.Open("mysql", sqlConnection)
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Answer{}, &AnswerHistory{}, &Message{}, &Question{}, &User{})
}

func setRouter(r martini.Router) {
	r.Post("/users", addUser)
	r.Get("/users/top", getUsersTop)
	r.Get("/users/:user_id", getUser)
	r.Post("/messages/slack", addMessage)
	r.Get("/messages", getMessages)
	r.Post("/slack/commands/tv", slackCommandTV)
}

func decodeRequestForm(r *http.Request, v interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return schema.NewDecoder().Decode(v, r.PostForm)
}

func decodeRequestQuery(r *http.Request, v interface{}) error {
	return schema.NewDecoder().Decode(v, r.URL.Query())
}

func main() {
	initDB()
	m := martini.Classic()
	setRouter(m.Router)
	m.Run()
}
