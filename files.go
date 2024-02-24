package gochat

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LoginReq struct {
}

type FileUpload struct {
	broker *Broker
	root   fs.FS
}

func NewFileUpload(broker *Broker) *FileUpload {
	return &FileUpload{
		broker: broker,
	}
}

func (f *FileUpload) Upload(channel string, msg, user string, file multipart.File, filename string) error {
	id := uuid.NewString()
	ext := path.Ext(filename)
	name := fmt.Sprintf("%s.%s", id, ext)

	dst, err := os.Create("./uploads/" + name)
	if err != nil {
		log.Println(err)
	}

	_, err = io.Copy(dst, file)
	if err != nil {
		log.Println(err)
	}

	ext = strings.ToLower(ext)

	if ext == ".jpg" || ext == ".png" || ext == ".jpeg" || ext == ".gif" {
		log.Println("file image upload")
		m := &Message{
			Channel:   channel,
			User:      user,
			Timestamp: time.Now().UnixMilli(),
			Content:   msg,
			Image:     fmt.Sprintf("/uploads/%s", name),
		}
		f.broker.Publish(m, true)
	} else {
		log.Println("attachment upload")
		file := &FileEntry{
			Filename: filename,
			URL:      fmt.Sprintf("/uploads/%s", name),
		}
		m := &Message{
			Channel:   channel,
			User:      user,
			Timestamp: time.Now().UnixMilli(),
			Content:   msg,
			File:      file,
		}

		f.broker.Publish(m, true)
	}

	return nil
}
