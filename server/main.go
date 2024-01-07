package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

type Message struct {
	GroupID string `json:"groupId"`
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

func main() {
	ctx := context.Background()

	nc, err := nats.Connect("nats:4222")
	if err != nil {
		panic(err)
	}

	slog.InfoContext(ctx, "connect to NATS server", slog.Bool("isConnected", nc.IsConnected()))

	r := chi.NewRouter()

	r.Post("/send", func(w http.ResponseWriter, r *http.Request) {
		var message Message
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			panic(err)
		}

		bytes, err := json.Marshal(message)
		if err != nil {
			panic(err)
		}

		if err := nc.Publish(message.GroupID, bytes); err != nil {
			panic(err)
		}
	})

	r.Get("/group/{groupID}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		groupID := chi.URLParam(r, "groupID")

		slog.InfoContext(ctx, "request join group", slog.String("group", groupID))

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		messages := make(chan Message)

		subscribe, err := nc.Subscribe(groupID, func(msg *nats.Msg) {
			var message Message
			if err := json.Unmarshal(msg.Data, &message); err != nil {
				panic(err)
			}
			messages <- message
		})
		if err != nil {
			panic(err)
		}
		defer func(subscribe *nats.Subscription) {
			err := subscribe.Unsubscribe()
			if err != nil {
				panic(err)
			}
		}(subscribe)

		for {
			select {
			case msg := <-messages:
				if _, err := w.Write([]byte("data: ")); err != nil {
					panic(err)
				}
				if err := json.NewEncoder(w).Encode(&msg); err != nil {
					panic(err)
				}
				if _, err := w.Write([]byte("\n\n")); err != nil {
					panic(err)
				}
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			case <-ctx.Done():
				slog.InfoContext(ctx, "client exit group", slog.String("group", groupID))
				return
			}
		}
	})

	slog.InfoContext(ctx, "starting server", slog.Int("port", 8080))
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
