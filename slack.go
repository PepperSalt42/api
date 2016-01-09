package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	commandTVUsage = "Valid commands: help, question, answer, image, status."
	commandTVFunc  = map[string]func(*SlackCommandRequest, *User) *SlackCommandResponse{
		"help":     slackCommandTVHelp,
		"question": slackCommandTVQuestion,
		"answer":   slackCommandTVAnswer,
		"status":   slackCommandTVStatus,
		"image":    slackCommandTVImage,
	}

	argsRegexp = regexp.MustCompile("'.+'|\".+\"|\\S+")
)

// SlackCommandRequest contains the data of slack command request.
type SlackCommandRequest struct {
	Token       string `schema:"token"`
	UserID      string `schema:"user_id"`
	Command     string `schema:"command"`
	Text        string `schema:"text"`
	ResponseURL string `schema:"response_url"`
}

// SlackCommandResponse contains the data of slack command request.
type SlackCommandResponse struct {
	Text string `json:"text"`
}

func slackCommandTV(w http.ResponseWriter, r *http.Request) {
	var req SlackCommandRequest
	if err := decodeRequestForm(r, &req); err != nil {
		renderJSON(w, http.StatusBadRequest, Error{err.Error()})
		return
	}
	if req.Token != slackOutgoingToken {
		renderJSON(w, http.StatusBadRequest, errInvalidToken)
		return
	}
	user, err := GetUserBySlackID(req.UserID)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, Error{err.Error()})
		return
	}
	resp, err := getCommandTVResponse(&req, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getCommandTVResponse(req *SlackCommandRequest, user *User) (*http.Response, error) {
	cmdStr := req.Text
	if i := strings.IndexRune(req.Text, ' '); i != -1 {
		cmdStr = cmdStr[:i]
	}
	var slackResp *SlackCommandResponse
	if cmdFunc, ok := commandTVFunc[cmdStr]; ok {
		slackResp = cmdFunc(req, user)
	} else {
		text := fmt.Sprintf("Invalid command %q.\n%s", cmdStr, commandTVUsage)
		slackResp = &SlackCommandResponse{Text: text}
	}
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(slackResp); err != nil {
		return nil, err
	}
	return http.Post(req.ResponseURL, ContentJSON, body)
}

func slackCommandTVHelp(req *SlackCommandRequest, user *User) *SlackCommandResponse {
	return &SlackCommandResponse{Text: commandTVUsage}
}

func slackCommandTVQuestion(req *SlackCommandRequest, user *User) *SlackCommandResponse {
	resp := &SlackCommandResponse{}
	if len(req.Text) <= len("question ") {
		resp.Text = commandTVUsage
		return resp
	}
	argsStr := req.Text[len("question "):]
	args := argsRegexp.FindAllString(argsStr, -1)
	if len(args) < 4 || len(args) > 6 {
		nbrAnswers := len(args) - 2
		if nbrAnswers < 0 {
			nbrAnswers = 0
		}
		resp.Text = fmt.Sprintf("Error: Can't set %d answers: Minimum 2 answers and maximum 4 answers", nbrAnswers)
		return resp
	}
	if len(args[0]) > 128 {
		resp.Text = "Error: Question is too long maximum 128 characters"
		return resp
	}
	answerIndex, _ := strconv.Atoi(args[1])
	answersStr := args[2:]
	if answerIndex <= 0 || answerIndex > len(answersStr) {
		resp.Text = "Error: Invalid right answer index"
		return resp
	}
	for i, answerStr := range answersStr {
		if len(answerStr) > 32 {
			resp.Text = fmt.Sprintf("Error: Answer %d is too long maximum 128 characters", i+1)
			return resp
		}
	}
	tx := db.Begin()
	question := &Question{UserID: user.ID, Sentence: args[0]}
	if err := tx.Create(question).Error; err != nil {
		tx.Rollback()
		resp.Text = fmt.Sprintf("Error: Can't create question: %s", err)
		return resp
	}
	for i, answerStr := range answersStr {
		answer := &Answer{QuestionID: question.ID, Sentence: answerStr}
		if err := tx.Create(answer).Error; err != nil {
			tx.Rollback()
			resp.Text = fmt.Sprintf("Error: Can't create answer: %s", err)
			return resp
		}
		if i == answerIndex-1 {
			if err := tx.Model(&question).UpdateColumn("right_answer_id", answer.ID).Error; err != nil {
				tx.Rollback()
				resp.Text = fmt.Sprintf("Error: Can't update question right_answer_id: %s", err)
				return resp
			}
		}
	}
	tx.Commit()
	resp.Text = "Your question has been submitted. Thank You!"
	return resp
}

func slackCommandTVAnswer(req *SlackCommandRequest, user *User) *SlackCommandResponse {
	resp := &SlackCommandResponse{}
	question, err := GetCurrentQuestion()
	if err != nil {
		resp.Text = fmt.Sprintf("Error: Can't get current question: %v", err)
		return resp
	}
	answers, err := GetAnswersByQuestionID(question.ID)
	if err != nil {
		resp.Text = fmt.Sprintf("Error: Can't get answers: %v", err)
		return resp
	}
	if len(req.Text) <= len("answer ") {
		resp.Text = commandTVUsage
		return resp
	}
	answerIndex, _ := strconv.Atoi(req.Text[len("answer "):])
	if answerIndex <= 0 || answerIndex > len(answers) {
		resp.Text = fmt.Sprintf("Invalid answer index.\nThere is %d possible answers.\nSee help and status for more details", len(answers))
		return resp
	}
	answer := answers[answerIndex-1]
	answerEntry := &AnswerEntry{UserID: user.ID, QuestionID: question.ID, AnswerID: answer.ID}
	if err := InsertOrUpdateDB(answerEntry, answerEntry); err != nil {
		resp.Text = fmt.Sprintf("Error: Can't add your answers: %v", err)
		return resp
	}
	resp.Text = fmt.Sprintf("Answer Added.\n%s %s", question.Sentence, answer.Sentence)
	return resp
}

func slackCommandTVStatus(req *SlackCommandRequest, user *User) *SlackCommandResponse {
	resp := &SlackCommandResponse{}
	question, err := GetCurrentQuestion()
	if err != nil {
		resp.Text = fmt.Sprintf("Error: Can't get current question: %v", err)
		return resp
	}
	questionUser, err := GetUser(question.UserID)
	if err != nil {
		resp.Text = fmt.Sprintf("Error: Can't get user associated to the current question: %v", err)
		return resp
	}
	answers, err := GetAnswersByQuestionID(question.ID)
	if err != nil {
		resp.Text = fmt.Sprintf("Error: Can't get answers: %v", err)
		return resp
	}
	topUsers, err := GetUsersTop(6)
	if err != nil {
		resp.Text = fmt.Sprintf("Error: Can't get top users: %v", err)
		return resp
	}
	buff := &bytes.Buffer{}
	fmt.Fprintf(buff, "Question from %s %s:\n%s\n", questionUser.FirstName, questionUser.LastName, question.Sentence)
	lastAnswerIndex := len(answers) - 1
	for i, answer := range answers {
		fmt.Fprintf(buff, "%d. %s", i+1, answer.Sentence)
		if i < lastAnswerIndex {
			buff.WriteString(", ")
		}
	}
	buff.WriteString("\n\nTop:\n")
	for _, user := range topUsers {
		fmt.Fprintf(buff, "%s %s: %v points\n", user.FirstName, user.LastName, user.Points)
	}
	resp.Text = buff.String()
	return resp
}

func slackCommandTVImage(req *SlackCommandRequest, user *User) *SlackCommandResponse {
	resp := &SlackCommandResponse{}
	if len(req.Text) <= len("image ") {
		resp.Text = commandTVUsage
		return resp
	}
	urlStr := req.Text[len("image "):]
	if _, err := url.Parse(urlStr); err != nil {
		resp.Text = "Error: Invalid image URL"
		return resp
	}
	if err := db.Create(&Image{URL: urlStr, UserID: user.ID}).Error; err != nil {
		resp.Text = fmt.Sprintf("Error: Can't add image to the database: %v", err)
		return resp
	}
	resp.Text = "Image added successfully!"
	return resp
}
