package main

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-martini/martini"
	"github.com/gorilla/schema"
	"github.com/martini-contrib/gzip"
)

const (
	// Default character encoding.
	defaultCharset = "; charset=UTF-8"
	// ContentJSON header value for JSON data.
	ContentJSON = "application/json" + defaultCharset
	// ContentType header constant.
	ContentType = "Content-Type"
)

var (
	errInvalidTimestamp    = Error{"Invalid timestamp"}
	errInvalidToken        = Error{"Invalid token"}
	errInvalidUserID       = Error{"Invalid user_id"}
	errMessagesNotFound    = Error{"Messages not found"}
	errUserNotFound        = Error{"User not found"}
	errCurQuestionNotFound = Error{"Current question not found"}
	errLastImageNotFound   = Error{"Last image not found"}
)

// Error exposes an error message
type Error struct {
	Error string `json:"error"`
}

// NewWebService creates a new web service ready to run.
func NewWebService() *martini.Martini {
	m := martini.New()
	m.Handlers(loggerMiddleware(), martini.Recovery(), gzip.All())
	r := newRouter()
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return m
}

// newRouter creates a router with all the web service routes set
func newRouter() martini.Router {
	r := martini.NewRouter()
	r.Get("/images/latest", getLastImage)
	r.Get("/users/top", getUsersTop)
	r.Get("/users/:user_id", getUser)
	r.Post("/messages/slack", addMessage)
	r.Get("/messages", getMessages)
	r.Get("/questions/current", getCurrentQuestion)
	r.Post("/slack/commands/tv", slackCommandTV)
	return r
}

// decodeRequestForm decodes a request form.
func decodeRequestForm(r *http.Request, v interface{}) error {
	if err := r.ParseForm(); err != nil {
		log.WithField("err", err).Info("Invalid request form")
		return err
	}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	return decoder.Decode(v, r.PostForm)
}

// decodeRequestQuery decodes a request form.
func decodeRequestQuery(r *http.Request, v interface{}) error {
	if err := schema.NewDecoder().Decode(v, r.URL.Query()); err != nil {
		log.WithField("err", err).Info("Invalid request query")
		return err
	}
	return nil
}

// renderJSON renders an object in json.
func renderJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set(ContentType, ContentJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.WithField("err", err).Error("Can't render to JSON")
	}
}

// loggerMiddleware is a martini middleware to log each request throw our logger.
func loggerMiddleware() martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		start := time.Now()
		addr := req.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = req.RemoteAddr
			}
		}
		rw := res.(martini.ResponseWriter)
		c.Next()
		log.WithFields(log.Fields{
			"method":      req.Method,
			"path":        req.URL.Path,
			"addr":        addr,
			"status":      rw.Status(),
			"status_text": http.StatusText(rw.Status()),
			"duration":    time.Since(start),
		}).Info("Completed")
	}
}
