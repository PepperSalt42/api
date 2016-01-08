package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var db gorm.DB

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

func main() {
	initDB()
	m := martini.Classic()
	m.Post("/users", addUser)
	m.Get("/users/:user_id", getUser)
	m.Run()
}
