package main

import (
	"github.com/jinzhu/gorm"
	"time"
)

// User contains all informations about an User
type User struct {
	gorm.Model
	SlackID    uint
	Name       string
	PictureURL string
	UpdatedAt  time.Time
	Points     uint
}
