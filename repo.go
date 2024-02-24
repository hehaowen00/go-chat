package gochat

import (
	"database/sql"

	"github.com/google/uuid"
)

type ChannelEntry struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) NewUser(username, password string) (string, error) {
	id := uuid.NewString()

	_, err := r.db.Exec("INSERT INTO users (user_id, username, password_hash) VALUES ($1, $2, $3)", id, username, password)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *Repo) NewChannel(name string) (string, error) {
	id := uuid.NewString()

	_, err := r.db.Exec("INSERT INTO channels (channel_id, name) VALUES ($1, $2)", id, name)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *Repo) GetChannels() ([]*ChannelEntry, error) {
	rows, err := r.db.Query("SELECT channel_id, name FROM channels")
	if err != nil {
		return nil, err
	}

	var channels []*ChannelEntry

	for rows.Next() {
		entry := ChannelEntry{}

		err = rows.Scan(&entry.Id, &entry.Name)
		if err != nil {
			return nil, err
		}

		channels = append(channels, &entry)
	}

	return channels, nil
}

func (r *Repo) GetChannel(id string) (string, error) {
	var name string
	err := r.db.QueryRow("SELECT name FROM channels WHERE channel_id = $1", id).Scan(&name)
	return name, err
}

func (r *Repo) GetChannelByName(name string) (string, error) {
	var id string
	err := r.db.QueryRow("SELECT channel_id FROM channels WHERE name = $1", name).Scan(&id)
	return id, err
}

func (r *Repo) SaveMessage(channel, user string, ts int64, message []byte) (string, error) {
	id := uuid.NewString()
	_, err := r.db.Exec("INSERT INTO messages (message_id, channel_id, user_id, timestamp, content) VALUES ($1, $2, $3, $4, $5)", id, channel, user, ts, message)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *Repo) GetMessages(channel string) ([][]byte, error) {
	rows, err := r.db.Query("SELECT content FROM messages WHERE channel_id = $1 LIMIT 100;", channel)
	if err != nil {
		return nil, err
	}

	var messages [][]byte

	for rows.Next() {
		var message []byte

		err = rows.Scan(&message)
		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	return messages, nil
}
