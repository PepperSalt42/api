package main

import (
	"github.com/jinzhu/gorm"
)

// Answer contains all informations about an Answer
type Answer struct {
	gorm.Model
	QuestionID uint
	Sentence   string
}

// GetAnswersByQuestionID returns the answers associated to the questionID.
func GetAnswersByQuestionID(questionID uint) (answers []Answer, err error) {
	err = db.Where(&Answer{QuestionID: questionID}).Find(&answers).Error
	return
}
