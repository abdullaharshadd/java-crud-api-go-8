package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"migrated-app/pkg/error"
	"migrated-app/pkg/service"
	"migrated-app/pkg/user"
	"migrated-app/pkg/repository"
	"migrated-app/pkg/db"
)

// UserController represents the API endpoints for user management.
type UserController struct {
	UserService service.UserService
}

// NewUserController creates a new instance of UserController.
func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

// SaveUser handles the POST /save_user_data endpoint.
func (uc *UserController) SaveUser(c *gin.Context) {
	var user user.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := uc.UserService.SaveUser(ctx, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User data saved successfully!"})
}

// FetchUserList handles the GET /get_user_data endpoint.
func (uc *UserController) FetchUserList(c *gin.Context) {
	ctx := context.Background()
	users, err := uc.UserService.FetchUserList(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// FetchUserByID handles the GET /get_user_data/{id} endpoint.
func (uc *UserController) FetchUserByID(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx := context.Background()
	user, err := uc.UserService.FetchUserByID(ctx, userID)
	if err != nil {
		if error.IsUserNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser handles the DELETE /delete_user_data/{id} endpoint.
func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx := context.Background()
	if err := uc.UserService.DeleteUser(ctx, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User data deleted successfully!"})
}

// UpdateUser handles the PUT /update_user_data/{id} endpoint.
func (uc *UserController) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var updatedUser user.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := uc.UserService.UpdateUser(ctx, userID, &updatedUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// GetUserByName handles the GET /get_user_name/name/{name} endpoint.
func (uc *UserController) GetUserByName(c *gin.Context) {
	name := c.Param("name")

	ctx := context.Background()
	user, err := uc.UserService.GetUserByName(ctx, name)
	if err != nil {
		if error.IsUserNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// SetupRoutes registers the user-related routes.
func SetupRoutes(r *gin.Engine, userService service.UserService) {
	userCtrl := NewUserController(userService)

	r.POST("/save_user_data", userCtrl.SaveUser)
	r.GET("/get_user_data", userCtrl.FetchUserList)
	r.GET("/get_user_data/:id", userCtrl.FetchUserByID)
	r.DELETE("/delete_user_data/:id", userCtrl.DeleteUser)
	r.PUT("/update_user_data/:id", userCtrl.UpdateUser)
	r.GET("/get_user_name/name/:name", userCtrl.GetUserByName)
}

// CreateUserTable ensures the user table exists in the database.
func CreateUserTable(db *db.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS ` + "`users`" + ` (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(255) NOT NULL,
		about TEXT
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create user table: %w", err)
	}
	return nil
}

// InitializeUserService initializes the UserService with a UserRepository.
func InitializeUserService(db *db.DB) service.UserService {
	userDAO := repository.NewUserDAO(db)
	return service.NewUserServiceImp(userDAO)
}

// InitializeAPI sets up the API routes and initializes the database schema.
func InitializeAPI(db *db.DB) *gin.Engine {
	r := gin.Default()

	if err := CreateUserTable(db); err != nil {
		log.Fatalf("Failed to create user table: %v", err)
	}

	userService := InitializeUserService(db)
	SetupRoutes(r, userService)

	return r
}