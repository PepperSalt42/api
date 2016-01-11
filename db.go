package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var db gorm.DB

// InitDB opens the database with the informations from the env.
// Automigrate all the tables.
func InitDB() {
	var err error
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	dbname := os.Getenv("MYSQL_DATABASE")
	sqlConnection := fmt.Sprintf("%s:%s@tcp(db:3306)/%s?charset=utf8&parseTime=True&loc=Local", user, password, dbname)
	db, err = gorm.Open("mysql", sqlConnection)
	if err != nil {
		log.WithFields(log.Fields{
			"user":     user,
			"password": password,
			"dbname":   dbname,
			"err":      err,
		}).Fatal("Can't open mysql database")
		log.Fatal(err)
	}
	db.AutoMigrate(&Answer{}, &AnswerEntry{}, &Image{}, &Message{}, &Question{}, &User{})
}

// InsertOrUpdateDB inserts or updates the values in the database.
func InsertOrUpdateDB(query, out interface{}) error {
	if db.Where(query).First(out).RecordNotFound() {
		return db.Create(out).Error
	}
	return db.Save(out).Error
}
