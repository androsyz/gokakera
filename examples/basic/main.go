package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/androsyz/gokakera"
	"github.com/androsyz/gokakera/localstorage"
	"github.com/androsyz/gokakera/memorystore"
)

func main() {
	storage := localstorage.New("./uploads")
	sessionStore := memorystore.New()

	kakera := gokakera.New(
		gokakera.WithChunkSize(5*1024*1024),
		gokakera.WithMaxFileSize(10*1024*1024*1024),
		gokakera.WithSessionTTL(24*time.Hour),
		gokakera.WithMaxConcurrency(5),
		gokakera.WithChecksumType("crc32"),
		gokakera.WithStorage(storage),
		gokakera.WithSessionStore(sessionStore),
		gokakera.WithOnProgress(func(p gokakera.Progress) {
			log.Printf("Session %s: %.1f%% (%d/%d chunks)",
				p.SessionID, p.Percentage, p.ChunksReceived, p.TotalChunks)
		}),
	)

	kakera.StartCleanup(context.Background(), 5*time.Minute)

	mux := http.NewServeMux()
	kakera.RegisterRoutes(mux)

	log.Println("gokakera server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
