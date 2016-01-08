package main

import (
	"bytes"
	"encoding/json"
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

func PostJSONRequest(t *testing.T, urlStr string, v interface{}) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(v); err != nil {
		t.Fatalf("Can't generate JSON payload: %v", err)
	}
	req := newRequest(t, "POST", urlStr, body)
	req.Header.Set(ContentType, ContentJSON)
	return DoRequest(req)
}

func TestMain(m *testing.M) {
	initDB()
	db.DropTable(&Answer{}, &AnswerHistory{}, &Message{}, &Question{}, &User{})
	db.CreateTable(&Answer{}, &AnswerHistory{}, &Message{}, &Question{}, &User{})
	mc = martini.Classic()
	setRouter(mc.Router)
	os.Exit(m.Run())
}

func TestUsers(t *testing.T) {
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
