// Package smartcontact holds the application bootstrap for the Smart Contact
// Manager. It replaces Spring Boot's @SpringBootApplication /
// SpringApplication.run. It loads configuration, applies command-line
// overrides, opens the data source, creates the schema, wires dependencies
// explicitly (instead of component scanning), builds the router and manages
// the HTTP server lifecycle.
package smartcontact

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"migrated-app/cmd/api/api"
	"migrated-app/internal/config"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/pkg/db"
	"migrated-app/pkg/service"
)

// defaultPort is the port used when none is configured. It matches the
// server.port default in Spring Boot.
const defaultPort = "8080"

// shutdownTimeout limits how long a graceful shutdown may take.
const shutdownTimeout = 10 * time.Second

// serverPortArg is the Spring Boot command-line property for the HTTP port.
const serverPortArg = "--server.port"

// createUserTableDDL creates the table that backs com.smartContact.model.User.
// The table and column names match the JPA @Table/@Column mappings exactly.
// They are the same identifiers that every query in
// internal/smartcontact/repository/user_dao.go uses.
// Every identifier is backtick-quoted for two reasons. USER is a reserved word
// in MySQL, and MySQL table names are case-sensitive on most platforms.
// MIGRATION_NOTE: string columns use VARCHAR(255), which is Hibernate's
// default for a String field with no @Column(length=...).
const createUserTableDDL = "CREATE TABLE IF NOT EXISTS `USER` (" +
	"`User_id` INT NOT NULL AUTO_INCREMENT, " +
	"`User_name` VARCHAR(255), " +
	"`User_Email` VARCHAR(255), " +
	"`User_Password` VARCHAR(255), " +
	"`User_Role` VARCHAR(255), " +
	"`User_About` VARCHAR(255), " +
	"PRIMARY KEY (`User_id`)" +
	")"

// Application is the explicitly wired equivalent of the Spring ApplicationContext.
type Application struct {
	db     *sql.DB
	server *http.Server
}

// New builds the Application. It loads configuration, applies command-line
// overrides from args, opens the database, creates the schema and wires every
// component by hand.
// MIGRATION_NOTE: Spring's classpath auto-configuration and component scanning
// have no Go equivalent. Every dependency is constructed explicitly here.
func New(ctx context.Context, args []string) (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load configuration: %w", err)
	}

	port := cfg.Port
	if override, ok := portFromArgs(args); ok {
		port = override
	}
	if port == "" {
		port = defaultPort
	}

	conn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Replaces Hibernate ddl-auto. Create the real schema at boot.
	if err := CreateSchema(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &Application{
		db: conn,
		server: &http.Server{
			Addr:              ":" + port,
			Handler:           NewHandler(conn),
			ReadHeaderTimeout: 10 * time.Second,
		},
	}, nil
}

// CreateSchema creates the tables required by the application if they do not
// already exist.
func CreateSchema(ctx context.Context, conn *sql.DB) error {
	if _, err := conn.ExecContext(ctx, createUserTableDDL); err != nil {
		return fmt.Errorf("create USER table: %w", err)
	}
	return nil
}

// NewHandler wires the repository, service and controllers onto a Gin engine.
// The repository comes from internal/smartcontact/repository. It queries the
// same `USER` table that CreateSchema creates.
func NewHandler(conn *sql.DB) http.Handler {
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	userDAO := repository.NewUserDAO(conn)
	userService := service.NewUserService(userDAO)
	api.RegisterRoutes(r.Group(""), userService)

	return r
}

// portFromArgs extracts a Spring-style --server.port override from args.
// It accepts both "--server.port=8081" and "--server.port 8081". Other
// arguments are ignored.
func portFromArgs(args []string) (string, bool) {
	port, found := "", false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, serverPortArg+"="):
			port, found = strings.TrimPrefix(arg, serverPortArg+"="), true
		case arg == serverPortArg && i+1 < len(args):
			i++
			port, found = args[i], true
		}
	}
	if port == "" {
		return "", false
	}
	return port, found
}

// Run starts the HTTP server and blocks until ctx is cancelled or the server
// fails. It then shuts down gracefully and releases the database.
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

// Run is the Go equivalent of SpringApplication.run(SmartContactApplication.class, args).
// It builds the application from args and starts it. The caller typically
// passes os.Args[1:] and a signal-aware context.
func Run(ctx context.Context, args []string) error {
	app, err := New(ctx, args)
	if err != nil {
		return err
	}
	return app.Run(ctx)
}
