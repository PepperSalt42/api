package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-martini/martini"
)

const (
	ContentFormURLEncoded = "application/x-www-form-urlencoded"
)

var mc *martini.Martini

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	InitDB()
	db.LogMode(true)
	initSlackServer()
	teardown()
	mc = NewWebService()
	os.Exit(m.Run())
}

func TestGetLastImage(t *testing.T) {
	defer teardown()
	imageURL := "http://localhost.com/image.png"
	userID := uint(1)
	db.Create(&Image{URL: imageURL, UserID: userID})
	resp := DoRequest(newRequest(t, "GET", "/images/latest", nil))
	if resp.Code != http.StatusOK {
		t.Fatal("Invalid status code:", resp.Code)
	}
	var img Image
	if err := json.NewDecoder(resp.Body).Decode(&img); err != nil {
		t.Fatal("Can't decode image:", err)
	}
	if img.ID != 1 || img.URL != imageURL || img.UserID != userID {
		t.Fatal("Invalid image:", img)
	}
}

func TestUsers(t *testing.T) {
	defer teardown()
	user := &User{FirstName: "John", LastName: "Doe"}
	db.Create(user)
	resp := DoRequest(newRequest(t, "GET", "/users/1", nil))
	if resp.Code != http.StatusOK {
		t.Fatal("Invalid status code:", resp.Code)
	}
	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		t.Fatal("Can't decode user:", err)
	}
	if u.ID != 1 || u.FirstName != "John" || u.LastName != "Doe" {
		t.Fatal("Invalid user:", u)
	}
}

func TestAddMessage(t *testing.T) {
	defer teardown()
	addTestMessage(t, "UD10923", "helloworld")
	u := &User{}
	if err := db.First(u, 1).Error; err != nil {
		t.Fatal("Can't get user:", err)
	}
	if u.ID != 1 || u.SlackID != "UD10923" || u.FirstName != "John" || u.LastName != "Doe" || u.ImageURL != "http://localhost/image.jpg" {
		t.Fatal("Invalid user:", u)
	}
	message := &Message{}
	if err := db.First(message, 1).Error; err != nil {
		t.Fatal("Can't get message:", err)
	}
	if message.ID != 1 || message.Message != "helloworld" {
		t.Fatal("Invalid message:", message)
	}
}

func TestGetMessages(t *testing.T) {
	defer teardown()
	for i := 0; i < 10; i++ {
		addTestMessage(t, "UD10923", "helloworld")
	}
	req := newRequest(t, "GET", "/messages", nil)
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Can't get messages:", resp.Code)
	}

	req = newRequest(t, "GET", "/messages?from_id=5&count=2", nil)
	resp = DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Can't get messages:", resp.Code)
	}
	var data []Message
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatal("Can't decode json:", err)
	}
	if len(data) > 2 || len(data) < 2 {
		t.Fatal("Wrong messages number:", len(data))
	}
	if data[0].ID != 10 {
		t.Fatal("Incorrect message get:", data[0])
	}
	if data[1].ID != 9 {
		t.Fatal("Incorrect message get:", data[1])
	}
}

func TestGetUsersTop(t *testing.T) {
	defer teardown()
	for i := 0; i < 10; i++ {
		newID := fmt.Sprintf("UD%d", i)
		user := &User{SlackID: newID, FirstName: "John", LastName: "Doe", Points: uint(i)}
		if err := db.Create(user).Error; err != nil {
			t.Fatal("Can't update user:", err.Error())
		}
	}
	req := newRequest(t, "GET", "/users/top", nil)
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Can't get users top:", resp.Code)
	}
	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		t.Fatal("Can't decode json:", err)
	}
	if len(users) != 6 {
		t.Fatal("Wrong users number:", len(users))
	}
}

func TestGetCurrentQuestion(t *testing.T) {
	defer teardown()
	db.Create(&Question{UserID: 1, Sentence: "Help?", RightAnswerID: 1})
	req := newRequest(t, "GET", "/questions/current", nil)
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Can't get current question:", resp)
	}
}

func TestNextQuestion(t *testing.T) {
	defer teardown()
	db.Create(&User{FirstName: "John", LastName: "Doe", Points: 0})
	db.Create(&Question{UserID: 1, Sentence: "Help?", RightAnswerID: 1, StartedAt: time.Now()})
	db.Create(&AnswerEntry{QuestionID: 1, AnswerID: 1, UserID: 1})
	db.Create(&Question{UserID: 1, Sentence: "Donation?", RightAnswerID: 2})
	if err := nextQuestion(); err != nil {
		t.Fatal("Can't execute next question:", err)
	}
	q, err := GetCurrentQuestion()
	if err != nil {
		t.Fatal("Can't get current question:", err)
	}
	if q.ID == 1 {
		t.Fatal("Invalid current question:", q)
	}
}

func TestSlackCommandHelp(t *testing.T) {
	defer teardown()
	params := fmt.Sprintf("token=%s&user_id=UD10923&command=tv&text=help&response_url=http://localhost:4242/commands/1234/5500", slackCommandToken)
	req := newRequest(t, "POST", "/slack/commands/tv", bytes.NewBufferString(params))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Invalid response:", resp.Code, resp.Body.String())
	}
}

func TestSlackCommandQuestion(t *testing.T) {
	defer teardown()
	bodies := []string{
		fmt.Sprintf("token=%s&user_id=UD10923&command=tv&text=question&response_url=http://localhost:4242/commands/1234/5600", slackCommandToken),
		fmt.Sprintf("token=%s&user_id=UD10923&command=tv&text=question Alive? 1 true false&response_url=http://localhost:4242/commands/1234/5601", slackCommandToken),
	}
	for _, bodyStr := range bodies {
		req := newRequest(t, "POST", "/slack/commands/tv", bytes.NewBufferString(bodyStr))
		req.Header.Set(ContentType, ContentFormURLEncoded)
		resp := DoRequest(req)
		if resp.Code != http.StatusOK {
			t.Fatal("Invalid response:", resp.Code, resp.Body.String())
		}
	}
}

func TestSlackCommandAnswer(t *testing.T) {
	defer teardown()
	db.Create(&Question{UserID: 1, Sentence: "Help?", RightAnswerID: 1})
	db.Create(&Answer{QuestionID: 1, Sentence: "Yes"})
	bodies := []string{
		fmt.Sprintf("token=%s&user_id=UD10923&command=tv&text=answer&response_url=http://localhost:4242/commands/1234/5700", slackCommandToken),
		fmt.Sprintf("token=%s&user_id=UD10923&command=tv&text=answer 1&response_url=http://localhost:4242/commands/1234/5701", slackCommandToken),
		fmt.Sprintf("token=%s&user_id=UD10923&command=tv&text=answer 2&response_url=http://localhost:4242/commands/1234/5702", slackCommandToken),
	}
	for _, bodyStr := range bodies {
		req := newRequest(t, "POST", "/slack/commands/tv", bytes.NewBufferString(bodyStr))
		req.Header.Set(ContentType, ContentFormURLEncoded)
		resp := DoRequest(req)
		if resp.Code != http.StatusOK {
			t.Fatal("Invalid response:", resp.Code, resp.Body.String())
		}
	}
}

func TestSlackCommandStatus(t *testing.T) {
	defer teardown()
	db.Create(&User{SlackID: "UD10923", FirstName: "John", LastName: "Doe", Points: 42})
	db.Create(&Question{UserID: 1, Sentence: "Help?", RightAnswerID: 1})
	db.Create(&Answer{QuestionID: 1, Sentence: "Yes"})
	db.Create(&Answer{QuestionID: 1, Sentence: "No"})
	params := fmt.Sprintf("token=%s&user_id=UD10923&command=tv&text=status&response_url=http://localhost:4242/commands/1234/5800", slackCommandToken)
	req := newRequest(t, "POST", "/slack/commands/tv", bytes.NewBufferString(params))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Invalid response:", resp.Code, resp.Body.String())
	}
}

func TestSlackCommandImage(t *testing.T) {
	defer teardown()
	db.Create(&User{SlackID: "UD10923", FirstName: "John", LastName: "Doe", Points: 42})
	imageURL := "http://localhost.com/image.png"
	params := fmt.Sprintf("token=%s&user_id=UD10923&command=tv&text=image %s&response_url=http://localhost:4242/commands/1234/5900", slackCommandToken, imageURL)
	req := newRequest(t, "POST", "/slack/commands/tv", bytes.NewBufferString(params))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Invalid response:", resp.Code, resp.Body.String())
	}
	img, err := GetLastImage()
	if err != nil {
		t.Fatal("Can't get last image:", err)
	}
	if img.URL != imageURL {
		t.Fatalf("Invalid image url in DB: %q != %q", img.URL, imageURL)
	}
	if img.UserID != 1 {
		t.Fatalf("Invalid image user_id in DB: %d != 1", img.UserID)
	}
}

func initSlackServer() {
	slackCommandToken = "legitCommandToken42"
	slackOutgoingToken = "legitOutgoingToken42"
	slackAPIToken = "legitAPIToken42"
	slackURL = "http://localhost:4242"
	m := martini.Classic()
	m.Get("/api/users.info", slackUserInfo)
	m.Post("/commands/1234/5500", slackCommandHandler(commandTVUsage))
	m.Post("/commands/1234/5600", slackCommandHandler(commandTVUsage))
	m.Post("/commands/1234/5601", slackCommandHandler("Your question has been submitted. Thank You!"))
	m.Post("/commands/1234/5700", slackCommandHandler(commandTVUsage))
	m.Post("/commands/1234/5701", slackCommandHandler("Answer Added.\nHelp? Yes"))
	m.Post("/commands/1234/5702", slackCommandHandler("Invalid answer index.\nThere is 1 possible answers.\nSee help and status for more details"))
	m.Post("/commands/1234/5800", slackCommandHandler("Question from John Doe:\nHelp?\n1. Yes, 2. No\n\nTop:\nJohn Doe: 42 points\n"))
	m.Post("/commands/1234/5900", slackCommandHandler("Image added successfully!"))
	go m.RunOnAddr(":4242")
}

func slackUserInfo(w http.ResponseWriter, r *http.Request) {
	renderJSON(w, http.StatusOK, struct {
		OK   bool      `json:"ok"`
		User SlackUser `json:"user"`
	}{
		OK: true,
		User: SlackUser{
			ID: "UD10923",
			Profile: SlackProfile{
				FirstName: "John",
				LastName:  "Doe",
				ImageURL:  "http://localhost/image.jpg",
			},
		},
	})
}
func slackCommandHandler(text string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if contentType := r.Header.Get(ContentType); contentType != ContentJSON {
			renderJSON(w, http.StatusBadRequest, fmt.Sprint("Invalid content type: ", contentType))
			return
		}
		var slackResp SlackCommandResponse
		if err := json.NewDecoder(r.Body).Decode(&slackResp); err != nil {
			renderJSON(w, http.StatusBadRequest, fmt.Sprint("Can't decode to JSON: ", err))
			return
		}
		if slackResp.Text != text {
			renderJSON(w, http.StatusBadRequest, fmt.Sprintf("Invalid response text: %q != %q", slackResp.Text, text))
			return
		}
		renderJSON(w, http.StatusOK, "OK")
	}
}

func newRequest(t *testing.T, method, urlStr string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		t.Fatalf("Can't create http request: %v", err)
	}
	return req
}

func DoRequest(req *http.Request) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	mc.ServeHTTP(resp, req)
	return resp
}

func teardown() {
	db.DropTable(&Answer{}, &AnswerEntry{}, &Image{}, &Message{}, &Question{}, &User{})
	db.CreateTable(&Answer{}, &AnswerEntry{}, &Image{}, &Message{}, &Question{}, &User{})
}

func addTestMessage(t *testing.T, userID string, text string) {
	params := fmt.Sprintf("token=%s&user_id=%s&text=%s&timestamp=%d", slackOutgoingToken, userID, text, time.Now().Unix())
	req := newRequest(t, "POST", "/messages/slack", bytes.NewBufferString(params))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp1 := DoRequest(req)
	if resp1.Code != http.StatusOK {
		t.Fatal("Message not added:", resp1.Code)
	}
}
