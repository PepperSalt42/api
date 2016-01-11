package main

import (
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	InitDB()
	go refreshQuestion()
	m := NewWebService()
	m.Run()
}
