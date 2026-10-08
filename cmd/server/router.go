```go
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"migrated-app/cmd/api"
	"migrated-app/pkg/db"
)

func buildRouter(dbConn *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	userService := api.InitializeUserService(dbConn)
	api.RegisterRoutes(r, userService)

	return r
}
```