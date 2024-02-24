package gochat

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WS struct {
}

func (ws *WS) Subscribe(w http.ResponseWriter, r *http.Request, b *Broker) {
	var upgrader websocket.Upgrader
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "error upgrading connection", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	token, err := r.Cookie("auth")
	if err != nil {
	}
	_ = token

	channel := r.PathValue("channel")
	user := ""

	log.Println(channel, "username", user)

	c, ok := b.channels[channel]
	if !ok {
		b.mu.Lock()
		c = &Channel{
			repo:    b.repo,
			clients: make(map[string]chan string),
		}

		_, err := b.repo.GetChannel(channel)
		if err != nil {
			log.Println(err)
			http.Error(w, "error channel not found", http.StatusInternalServerError)
			b.mu.Unlock()
			return
		}

		messages, err := b.repo.GetMessages(channel)
		if err != nil {
			log.Println(err)
		}
		c.history = messages

		b.channels[channel] = c

		b.mu.Unlock()
	}

	out := make(chan string, 10)
	c.mu.Lock()
	connID := uuid.NewString()
	c.clients[connID] = out
	c.mu.Unlock()

	for _, msg := range c.history {
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
		w.(http.Flusher).Flush()
	}
	log.Println("history sent")

	b.Publish(&Message{
		Type:    MessageType_System,
		Channel: channel,
		User:    "",
		Content: fmt.Sprintf("user %s joined", user),
	}, false)

	for {
		select {
		case msg := <-out:
			conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
}
