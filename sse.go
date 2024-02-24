package gochat

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type SSE struct {
}

func (sse *SSE) Subscribe(w http.ResponseWriter, r *http.Request, b *Broker) {
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

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fmt.Fprintf(w, "event: status\ndata: joined\n\n")
	w.(http.Flusher).Flush()

	out := make(chan string, 10)

	c.mu.Lock()
	conn := uuid.NewString()
	c.clients[conn] = out
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
		case <-r.Context().Done():
			log.Println("unsubscribed", user)
			c.mu.Lock()
			close(out)
			delete(c.clients, user)
			c.mu.Unlock()

			b.Publish(&Message{
				Type:    MessageType_System,
				Channel: channel,
				User:    "",
				Content: fmt.Sprintf("user %s disconnected", user),
			}, false)
			return
		case msg := <-out:
			if !ok {
				return
			}

			n, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			log.Println("sent", n, err)
			w.(http.Flusher).Flush()
		}
	}
}
