package main

import "github.com/jinzhu/gorm"

// AnswerEntry contains all informations about an AnswerEntry
type AnswerEntry struct {
	gorm.Model
	UserID     uint
	QuestionID uint
	AnswerID   uint
}
