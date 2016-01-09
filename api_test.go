package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-martini/martini"
)

const (
	ContentFormURLEncoded = "application/x-www-form-urlencoded"
)

var mc *martini.ClassicMartini

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

func slackUserInfo(w http.ResponseWriter, r *http.Request) {
	renderJSON(w, http.StatusOK, struct {
		OK   bool      `json:"ok"`
		User SlackUser `json:"user"`
	}{
		OK: true,
		User: SlackUser{
			ID:   "UD10923",
			Name: "name",
			Profile: SlackProfile{
				ImageURL: "http://localhost/image.jpg",
			},
		},
	})
}

func initSlackServer() {
	slackOutgoingToken = "legitOutgoingToken42"
	slackAPIToken = "legitAPIToken42"
	slackURL = "http://localhost:4242"
	m := martini.Classic()
	m.Get("/api/users.info", slackUserInfo)
	m.Post("/commands/1234/5678", slackCommandHandler(commandTVUsage))
	m.Post("/commands/1234/5679", slackCommandHandler(commandTVUsage))
	m.Post("/commands/1234/5680", slackCommandHandler(commandTVUsage))
	m.Post("/commands/1234/5681", slackCommandHandler(commandTVUsage))
	go m.RunOnAddr(":4242")
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
	db.DropTable(&Answer{}, &AnswerHistory{}, &Message{}, &Question{}, &User{})
	db.CreateTable(&Answer{}, &AnswerHistory{}, &Message{}, &Question{}, &User{})
}

func TestMain(m *testing.M) {
	initDB()
	initSlackServer()
	teardown()
	mc = martini.Classic()
	setRouter(mc.Router)
	os.Exit(m.Run())
}

func TestUsers(t *testing.T) {
	defer teardown()
	req := newRequest(t, "POST", "/users", bytes.NewBufferString("name=toto"))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp1 := DoRequest(req)
	if resp1.Code != http.StatusCreated {
		t.Fatal("User not created:", resp1.Code)
	}
	resp2 := DoRequest(newRequest(t, "GET", "/users/1", nil))
	u := &User{}
	if resp2.Code != http.StatusOK {
		t.Fatal("Invalid status code:", resp1.Code)
	}
	if err := json.NewDecoder(resp2.Body).Decode(u); err != nil {
		t.Fatal("Can't decode user:", err)
	}
	if u.ID != 1 || u.Name != "toto" {
		t.Fatal("Invalid user:", u)
	}
}

func TestAddMessage(t *testing.T) {
	defer teardown()
	params := fmt.Sprintf("token=%s&user_id=U2147483697&text=bot: helloworld", slackOutgoingToken)
	req := newRequest(t, "POST", "/messages", bytes.NewBufferString(params))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp1 := DoRequest(req)
	if resp1.Code != http.StatusOK {
		t.Fatal("Message not added:", resp1.Code)
	}
	user := &User{}
	if err := db.First(user, 1).Error; err != nil {
		t.Fatal("Can't get user:", err)
	}
	if user.ID != 1 || user.SlackID != "UD10923" || user.Name != "name" || user.ImageURL != "http://localhost/image.jpg" {
		t.Fatal("Invalid user:", user)
	}
	message := &Message{}
	if err := db.First(message, 1).Error; err != nil {
		t.Fatal("Can't get message:", err)
	}
	if message.ID != 1 || message.Message != "bot: helloworld" {
		t.Fatal("Invalid message:", message)
	}
}

func TestSlackCommandHelp(t *testing.T) {
	defer teardown()
	params := fmt.Sprintf("token=%s&user_id=UD10923$command=tv&text=help&response_url=http://localhost:4242/commands/1234/5678", slackOutgoingToken)
	req := newRequest(t, "POST", "/slack/commands/tv", bytes.NewBufferString(params))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Invalid response:", resp.Code, resp.Body.String())
	}
}

func TestSlackCommandQuestion(t *testing.T) {
	defer teardown()
	params := fmt.Sprintf("token=%s&user_id=UD10923$command=tv&text=question&response_url=http://localhost:4242/commands/1234/5679", slackOutgoingToken)
	req := newRequest(t, "POST", "/slack/commands/tv", bytes.NewBufferString(params))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Invalid response:", resp.Code, resp.Body.String())
	}
}

func TestSlackCommandAnswer(t *testing.T) {
	defer teardown()
	params := fmt.Sprintf("token=%s&user_id=UD10923$command=tv&text=answer&response_url=http://localhost:4242/commands/1234/5680", slackOutgoingToken)
	req := newRequest(t, "POST", "/slack/commands/tv", bytes.NewBufferString(params))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Invalid response:", resp.Code, resp.Body.String())
	}
}

func TestSlackCommandStatus(t *testing.T) {
	defer teardown()
	params := fmt.Sprintf("token=%s&user_id=UD10923$command=tv&text=status&response_url=http://localhost:4242/commands/1234/5681", slackOutgoingToken)
	req := newRequest(t, "POST", "/slack/commands/tv", bytes.NewBufferString(params))
	req.Header.Set(ContentType, ContentFormURLEncoded)
	resp := DoRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatal("Invalid response:", resp.Code, resp.Body.String())
	}
}
