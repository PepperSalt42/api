package main

import (
	"github.com/jinzhu/gorm"
	"time"
)

// Message contains all informations about a Message
type Message struct {
	gorm.Model
	UserID   uint
	Message  string
	PostedAt time.Time
}
