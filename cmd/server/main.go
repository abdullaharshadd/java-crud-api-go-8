package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"migrated-app/internal/config"
	usermodel "migrated-app/internal/smartcontact/model"
	"migrated-app/pkg/db"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	dbConn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer dbConn.Close()

	// Create the USER table and the `users` compatibility view queried by
	// pkg/repository. Bounded so boot can never hang on an unreachable DB.
	schemaCtx, schemaCancel := context.WithTimeout(ctx, 30*time.Second)
	for {
		err = usermodel.EnsureSchema(schemaCtx, dbConn)
		if err == nil || schemaCtx.Err() != nil {
			break
		}
		log.Printf("schema not ready yet: %v", err)
		time.Sleep(time.Second)
	}
	schemaCancel()
	if err != nil {
		log.Printf("warning: ensure schema failed: %v", err)
	}

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	r := buildRouter(dbConn)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Printf("server started on :%s", port)
	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
