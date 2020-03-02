package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	//registerGreetingMenu()
	r := mux.NewRouter()
	r.HandleFunc("/", chatbotHandler)

	if err := http.ListenAndServe(":1369", r); err != nil {
		log.Fatal(err.Error())
	}
}

func chatbotHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		verifyWebhook(w, r)
	case "POST":
		processWebhook(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Don't suppor method %v", r.Method)
	}
}

func verifyWebhook(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	challenge := r.URL.Query().Get("hub.challenge")
	token := r.URL.Query().Get("hub.verify_token")

	if mode == "subscribe" && token == "1234" {
		w.WriteHeader(200)
		w.Write([]byte(challenge))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Error, wrong validation token"))
	}
}

func processWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Message not supported"))
		return
	}

	if req.Object == "page" {
		for _, entry := range req.Entry {
			for _, event := range entry.Messaging {
				if event.Message != nil {
					processMessage(&event)
				} else if event.PostBack != nil {
					processPostBack(&event)
				}
			}
		}

		w.WriteHeader(200)
		w.Write([]byte("Got your message"))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Message not supported"))
	}
}
