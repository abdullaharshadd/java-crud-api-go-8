package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"migrated-app/cmd/api/api"
	"migrated-app/pkg/repository"
	"migrated-app/pkg/service"
)

func buildRouter(dbConn *sql.DB) http.Handler {
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	userDAO := repository.NewUserDAO(dbConn)
	userService := service.NewUserServiceImp(userDAO)
	api.RegisterRoutes(r.Group(""), userService)

	return r
}
