package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"corematch/backend/internal/app"
)

func main() {
	coreApp := app.New()
	mux := http.NewServeMux()
	coreApp.RegisterRoutes(mux)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	coreApp.StartSnapshotLoop(ctx, 100*time.Millisecond)
	coreApp.StartConsoleLoop(ctx, 5*time.Second)

	addr := ":8080"
	log.Printf("corematch backend listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
