package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"
)

type ActionFunc func(chatId int64) error

type webhookReqBody struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		From struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"from"`
	} `json:"message"`
}

type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

var token string
var usersAllowList []string

var actionsHandler = map[string]ActionFunc{
	"marco": marcoPolo,
}

var chatsToNotify []int64

func main() {
	load_env()
	http.ListenAndServe(":8080", http.HandlerFunc(handler))
}

func load_env() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token = os.Getenv("TOKEN")
	usersAllowListUnparsed := os.Getenv("USERS_ALLOW_LIST")
	usersAllowList = strings.Split(usersAllowListUnparsed, ",")
}

func handler(w http.ResponseWriter, r *http.Request) {
	body := &webhookReqBody{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		fmt.Println("could not decode request body", err)
		return
	}

	if !contains(usersAllowList, body.Message.From.Username) {
		fmt.Println("user not allowed")
	}

	handler := actionsHandler[strings.ToLower(body.Message.Text)]
	if handler == nil {
		fmt.Println("no handler for this action")
		return
	}

	if err := handler(body.Message.Chat.ID); err != nil {
		fmt.Println("error in sending reply:", err)
		return
	}

	fmt.Println("reply sent")

}

func messageSender(chatId int64, message string) error {
	reqBody := &sendMessageReqBody{
		ChatID: chatId,
		Text:   message,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	res, err := http.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status" + res.Status)
	}

	return nil
}

func marcoPolo(chatId int64) error {
	chatsToNotify = append(chatsToNotify, chatId)
	return messageSender(chatId, "Polo!!")
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
