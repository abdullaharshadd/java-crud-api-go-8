// Package smartcontact holds the application bootstrap for the Smart Contact
// Manager. It replaces Spring Boot's @SpringBootApplication /
// SpringApplication.run: configuration loading, data source creation, schema
// creation, explicit dependency wiring (in place of component scanning),
// router construction and the HTTP server lifecycle.
package smartcontact

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"migrated-app/cmd/api/api"
	"migrated-app/internal/config"
	"migrated-app/pkg/db"
	"migrated-app/pkg/repository"
	"migrated-app/pkg/service"
	"migrated-app/pkg/user"
)

// defaultPort is the port used when none is configured (Spring Boot's server.port default).
const defaultPort = "8080"

// shutdownTimeout limits how long a graceful shutdown may take.
const shutdownTimeout = 10 * time.Second

// Application is the explicitly wired equivalent of the Spring ApplicationContext.
type Application struct {
	cfg    *config.Config
	db     *sql.DB
	server *http.Server
}

// New builds the Application: it loads configuration, opens the database,
// creates the schema and wires every component by hand.
// MIGRATION_NOTE: Spring's classpath auto-configuration and component scanning
// have no Go equivalent; every dependency is constructed explicitly here.
func New(ctx context.Context) (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load configuration: %w", err)
	}

	conn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Replaces Hibernate ddl-auto: create the real schema at boot.
	if err := user.CreateUserTable(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	port := cfg.Port
	if port == "" {
		port = defaultPort
	}

	return &Application{
		cfg: cfg,
		db:  conn,
		server: &http.Server{
			Addr:              ":" + port,
			Handler:           NewHandler(conn),
			ReadHeaderTimeout: 10 * time.Second,
		},
	}, nil
}

// NewHandler wires repositories, services and controllers onto a Gin engine.
func NewHandler(conn *sql.DB) http.Handler {
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	userDAO := repository.NewUserDAO(conn)
	userService := service.NewUserServiceImp(userDAO)
	api.RegisterRoutes(r.Group(""), userService)

	return r
}

// Run starts the HTTP server and blocks until ctx is cancelled or the server
// fails, then shuts down gracefully and releases the database.
func (a *Application) Run(ctx context.Context) error {
	defer func() {
		if err := a.db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("server started on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err, ok := <-errCh:
		if ok && err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := a.server.Shutdown(shutCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	return nil
}

// Run is the Go equivalent of SpringApplication.run: build and start the app.
func Run(ctx context.Context) error {
	app, err := New(ctx)
	if err != nil {
		return err
	}
	return app.Run(ctx)
}
