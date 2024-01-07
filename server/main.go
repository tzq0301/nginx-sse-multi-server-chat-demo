package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

type Message struct {
	SenderID string `json:"senderId"`
	Content  string `json:"message"`
}

func main() {
	ctx := context.Background()

	nc, err := nats.Connect("nats:4222")
	if err != nil {
		panic(err)
	}

	slog.InfoContext(ctx, "setup NATS server", slog.Bool("isConnected", nc.IsConnected()))

	r := chi.NewRouter()

	r.Get("/group/{groupID}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		groupID := chi.URLParam(r, "groupID")

		slog.InfoContext(ctx, "request join group", slog.String("group", groupID))

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		for {
			select {
			case <-time.After(1 * time.Second):
				if _, err := w.Write([]byte("data: ")); err != nil {
					panic(err)
				}
				if err := json.NewEncoder(w).Encode(&Message{
					Content:  "Hello",
					SenderID: "World",
				}); err != nil {
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
