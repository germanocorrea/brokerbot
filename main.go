package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"golang.ngrok.com/ngrok/v2"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

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

type ActionFunc func(message webhookReqBody) error

type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

var token string
var password string
var socketPath string
var useNgrok bool
var address string

var actionsHandler = map[string]ActionFunc{
	"marco": marcoPolo,
}

var chatsToNotify []int64
var wg sync.WaitGroup

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wg.Add(2)
	loadFlags()
	if useNgrok && os.Getenv("NGROK_AUTHTOKEN") == "" {
		log.Fatal("Error: NGROK_AUTHTOKEN environment variable is not set")
	}

	go func() {
		defer wg.Done()
		systemMessageBroker(ctx)
	}()

	go func() {
		if err := startWebhook(ctx); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, net.ErrClosed) {
			log.Printf("Webhook server error: %v", err)
		}
	}()

	log.Println("Bot started, press Ctrl+C to stop")
	<-ctx.Done()

	log.Println("Shutting services down...")
	wg.Wait()
	log.Println("Done")

}

func loadFlags() {
	tokenFlag := flag.String("token", "", "Telegram bot token")
	passwordFlag := flag.String("password", "", "Password to interact with the bot")
	socketFlag := flag.String("socketPath", "", "Path to the unix socketPath to listen to")
	ngrokFlag := flag.Bool("ngrok", false, "Use ngrok for the webhook")
	addressFlag := flag.String("address", "", "Webhook address to listen to")

	flag.Parse()

	token = *tokenFlag
	if token == "" {
		log.Fatal("Error: TOKEN environment variable is not set")
	}
	password = *passwordFlag
	if password == "" {
		log.Println("Warning: password not set, all users will be allowed to interact with the bot")
	}
	socketPath = *socketFlag
	if socketPath == "" {
		socketPath = getSocketPath()
	}

	address = *addressFlag
	if address == "" {
		address = ":8080"
	}

	useNgrok = *ngrokFlag
}

func startWebhook(ctx context.Context) error {
	err := error(nil)
	if !useNgrok {
		err = serveStandard()
	} else {
		err = serveNgrok(ctx)
	}

	if ctx.Err() != nil {
		log.Println("Webhook server stopped")
		return nil
	}
	return err
}

func serveStandard() error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	if err := setWebhook(address); err != nil {
		ln.Close()
		return err
	}
	return http.Serve(ln, http.HandlerFunc(handler))
}

func serveNgrok(ctx context.Context) error {
	ln, err := ngrok.Listen(ctx)
	if err != nil {
		return err
	}
	address = ln.URL().String()
	if err := setWebhook(address); err != nil {
		ln.Close()
		return err
	}
	return http.Serve(ln, http.HandlerFunc(handler))
}

func setWebhook(address string) error {
	base, err := url.Parse("https://api.telegram.org/bot" + token + "/setWebhook")
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Add("url", address)
	base.RawQuery = params.Encode()
	res, err := http.Get(base.String())
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status " + res.Status)
	}
	log.Println("Webhook set successfully")
	return nil
}

func handler(_ http.ResponseWriter, r *http.Request) {
	body := &webhookReqBody{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Println("Could not decode request body:", err)
		return
	}

	if !contains(chatsToNotify, body.Message.Chat.ID) {
		if err := authHandler(*body); err != nil {
			log.Println("Bot auth error:", err)
			return
		}
	}

	handler := actionsHandler[strings.ToLower(body.Message.Text)]
	if handler == nil {
		return
	}

	if err := handler(*body); err != nil {
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

func authHandler(message webhookReqBody) error {
	if password != "" && message.Message.Text != password {
		return errors.New("Unauthorized")
	}
	chatsToNotify = append(chatsToNotify, message.Message.Chat.ID)
	if password != "" {
		return messageSender(message.Message.Chat.ID, "Authorized")
	}

	return nil
}

func marcoPolo(message webhookReqBody) error {
	chatId := message.Message.Chat.ID
	return messageSender(chatId, "Polo!!")
}

func contains[T comparable](s []T, e T) bool {
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

func systemMessageBroker(ctx context.Context) {
	defer wg.Done()

	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()
		socket.Close()
		if err := os.Remove(socketPath); err != nil {
			log.Println("Error removing unix socket:", err)
		}
	}()

	for {
		conn, err := socket.Accept()
		if err != nil {
			if ctx.Err() != nil {
				log.Println("System message broker stopped")
				return
			}
			log.Fatal(err)
		}
		go newMessageConnection(conn)
	}
}
