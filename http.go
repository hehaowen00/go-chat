package gochat

import (
	"encoding/json"
	"log"
	"sync"
)

const (
	MessageType_System = "system"
	MessageType_Msg    = "msg"
)

type Message struct {
	Type      string     `json:"type"`
	Channel   string     `json:"channel"`
	User      string     `json:"user"`
	Content   string     `json:"content"`
	Image     string     `json:"image"`
	File      *FileEntry `json:"file"`
	Timestamp int64      `json:"timestamp"`
}

type FileEntry struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

type Channel struct {
	clients map[string]chan string
	history [][]byte
	repo    *Repo
	mu      sync.Mutex
}

func (c *Channel) Publish(m *Message, store bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}

	if store {
		c.history = append(c.history, data)
		log.Println(c.repo.SaveMessage(m.Channel, m.User, m.Timestamp, data))
	}

	wg := sync.WaitGroup{}

	for _, out := range c.clients {
		wg.Add(1)
		go func(out chan string) {
			defer func() {
				recover()
				wg.Done()
			}()
			out <- string(data)
		}(out)
	}

	wg.Wait()
}
