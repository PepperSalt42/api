package main

import "github.com/jinzhu/gorm"

// AnswerHistory contains all informations about an AnswerHistory
type AnswerHistory struct {
	gorm.Model
	PlayerID   int
	QuestionID int
}
