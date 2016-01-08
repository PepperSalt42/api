package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Message contains all informations about a Message
type Message struct {
	gorm.Model
	UserID   uint
	Message  string
	PostedAt time.Time
}
