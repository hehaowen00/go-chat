package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	gochat "github.com/hehaowen00/go-chat"
	pathrouter "github.com/hehaowen00/path-router"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "chat.db?journal_mode=wal")
	if err != nil {
		panic(err)
	}

	err = gochat.RunMigrations(db)
	if err != nil {
		panic(err)
	}

	repo := gochat.NewRepo(db)
	broker := gochat.NewBroker(repo)
	fileUpload := gochat.NewFileUpload(broker)

	err = os.MkdirAll("uploads", 0755)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("starting server")

	router := pathrouter.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	router.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		req := gochat.LoginReq{}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Println(err)
			http.Error(w, "error decoding request", http.StatusBadRequest)
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: uuid.New().String(),
		})
	})

	dir := http.Dir("uploads/")
	router.Get("/uploads/:fileID", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fileID := r.PathValue("fileID")
		f, err := dir.Open(fileID)
		if err != nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		io.Copy(w, f)
	})

	router.Get("/channels", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		channels, err := repo.GetChannels()
		if err != nil {
			log.Println(err)
			http.Error(w, "error getting channels", http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(channels)
		if err != nil {
			log.Println(err)
			http.Error(w, "error getting channels", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(data)
	})

	router.Post("/channels/new", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		broker.NewChannel(w, r)
	})

	router.Get("/chat/:channel", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		broker.Subscribe(w, r, &gochat.WS{})
	})

	router.Post("/send/:channel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseMultipartForm(10 * 1024 * 1024)

		// cookie, err := r.Cookie("token")
		// if err != nil {
		// 	log.Println(err)
		// 	http.Error(w, "unauthorized", http.StatusUnauthorized)
		// 	return
		// }
		// auth := cookie.Value
		// log.Println("token", auth)
		auth := "user"

		channel := r.PathValue("channel")
		msg := r.FormValue("message")

		f, h, err := r.FormFile("upload")
		if err != nil {
			log.Println(err)
			if msg != "" {
				m := &gochat.Message{
					Channel:   channel,
					User:      auth,
					Content:   msg,
					Timestamp: time.Now().UnixMilli(),
					Type:      gochat.MessageType_Msg,
				}
				broker.Publish(m, true)
			}
		} else {
			if f != nil {
				fileUpload.Upload(channel, msg, auth, f, h.Filename)
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	http.ListenAndServe(":8080", router)
}
