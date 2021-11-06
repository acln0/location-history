package main

import (
	"context"
	"fmt"
	"github.com/abdoub/location-history/handler"
	"github.com/abdoub/location-history/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {

	port := os.Getenv("HISTORY_SERVER_LISTEN_ADDR")
	if port == "" {
		port = "8080"
	}

	ttl := os.Getenv("LOCATION_HISTORY_TTL_SECONDS")
	if ttl == "" {
		ttl = "30"
	}

	storeTTL, err := strconv.Atoi(ttl)
	if err != nil {
		log.Fatal(err)
	}

	log := log.New(os.Stdout, "", log.LstdFlags)
	s := store.NewHistoryStore(log, storeTTL)
	h := handler.NewHistoryHandler(log, s)

	mux := http.NewServeMux()
	mux.Handle(h.Dispatch("/v1/location/history/"))

	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Unable to start HTTP server: %v", err)
		}
	}()

	log.Printf("Server listening on port %s", port)

	var stopSig = make(chan os.Signal, 1)
	signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	select {
	case <-stopSig:
		log.Println("received signal for graceful server shutdown")
		if err := srv.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}

}