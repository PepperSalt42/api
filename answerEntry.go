package main

import "github.com/jinzhu/gorm"

// AnswerEntry contains information about an answer entry.
type AnswerEntry struct {
	gorm.Model
	UserID     uint
	QuestionID uint
	AnswerID   uint
}
