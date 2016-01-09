package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

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

	errInvalidToken        = Error{"Invalid token"}
	errInvalidUserID       = Error{"Invalid user_id"}
	errMessagesNotFound    = Error{"Messages not found"}
	errUserNotFound        = Error{"User not found"}
	errCurQuestionNotFound = Error{"Current question not found"}
	errLastImageNotFound   = Error{"Last image not found"}
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
	db.AutoMigrate(&Answer{}, &AnswerEntry{}, &Image{}, &Message{}, &Question{}, &User{})
}

// InsertOrUpdateDB inserts or updates the value in the database.
func InsertOrUpdateDB(query, out interface{}) error {
	if db.Where(query).First(out).RecordNotFound() {
		return db.Create(out).Error
	}
	return db.Save(out).Error
}

func setRouter(r martini.Router) {
	r.Get("/images/latest", getLastImage)
	r.Get("/users/top", getUsersTop)
	r.Get("/users/:user_id", getUser)
	r.Post("/messages/slack", addMessage)
	r.Get("/messages", getMessages)
	r.Get("/questions/current", getCurrentQuestion)
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

func refreshQuestion() {
	questionRefreshRate := os.Getenv("QUESTION_REFRESH_RATE")
	wait, err := time.ParseDuration(questionRefreshRate)
	if err != nil {
		log.Fatal("Can't convert question refresh rate")
	}
	for {
		time.Sleep(wait)
		if err := nextQuestion(); err != nil {
			log.Printf("[ERROR] Can't set nextQuestion: %v", err)
		}
	}
}

func main() {
	initDB()
	rand.Seed(time.Now().Unix())
	m := martini.Classic()
	setRouter(m.Router)
	go refreshQuestion()
	m.Run()
}
