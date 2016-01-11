package main

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

// Image contains the data about an image.
type Image struct {
	gorm.Model
	UserID uint
	URL    string
}

// getLastImage returns the last image.
func getLastImage(w http.ResponseWriter, r *http.Request) {
	img, err := GetLastImage()
	if err != nil {
		renderJSON(w, http.StatusNotFound, errLastImageNotFound)
		return
	}
	renderJSON(w, http.StatusOK, img)
}

// GetLastImage returns the last image in database.
func GetLastImage() (*Image, error) {
	img := &Image{}
	err := db.Last(img).Error
	return img, err
}
