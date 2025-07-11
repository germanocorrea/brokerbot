package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/joho/godotenv"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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

var wg sync.WaitGroup

func main() {
	wg.Add(1)
	load_env()
	go systemMessageBroker()
	err := http.ListenAndServe(":8080", http.HandlerFunc(handler))
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}

func load_env() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token = os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("Error: TOKEN environment variable is not set")
	}

	usersAllowListUnparsed := os.Getenv("USERS_ALLOW_LIST")
	if usersAllowListUnparsed == "" {
		log.Println("Warning: USERS_ALLOW_LIST not set, all users will be denied")
	}
	usersAllowList = strings.Split(usersAllowListUnparsed, ",")
}

func handler(_ http.ResponseWriter, r *http.Request) {
	body := &webhookReqBody{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Println("Could not decode request body:", err)
		return
	}

	if !contains(usersAllowList, body.Message.From.Username) {
		return
	}

	handler := actionsHandler[strings.ToLower(body.Message.Text)]
	if handler == nil {
		return
	}

	if err := handler(body.Message.Chat.ID); err != nil {
		log.Println("Error in sending reply:", err)
		return
	}

	log.Println("Reply sent successfully")
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

func getSocketPath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		return filepath.Join(runtimeDir, "brokerbot.sock")
	}
	return "/tmp/brokerbot.sock"
}

func newMessageConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	message := buf[:n]

	for i := range chatsToNotify {
		err := messageSender(chatsToNotify[i], string(message))
		if err != nil {
			log.Fatal(err)
		}
	}

}

func systemMessageBroker() {
	defer wg.Done()

	socket, err := net.Listen("unix", getSocketPath())
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := os.Remove(getSocketPath())
		if err != nil {
			log.Println("Error removing socket:", err)
		}
		os.Exit(1)
	}()

	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go newMessageConnection(conn)
	}
}
