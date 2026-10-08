package api

import (
	"context"
	"encoding/json"
	"net/http"

	"migrated-app/internal/logger"
	"migrated-app/pkg/error"
	"migrated-app/pkg/service"
	"migrated-app/pkg/user"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

// UserController represents the REST controller for user-related operations.
type UserController struct {
	UserService service.UserService
}

// NewUserController creates a new instance of UserController.
func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

// RegisterRoutes registers all the user-related routes with the given router.
func RegisterRoutes(router *gin.RouterGroup, userService service.UserService) {
	controller := NewUserController(userService)

	router.POST("/save_user_data", controller.saveUser)
	router.GET("/get_user_data", controller.fetchUserList)
	router.GET("/get_user_data/:id", controller.fetchUserById)
	router.DELETE("/delete_user_data/:id", controller.deleteUser)
	router.PUT("/update_user_data/:id", controller.updateUser)
	router.GET("/get_user_name/name/:name", controller.getUserNameByName)
}

// saveUser handles the POST request to save a user entity to the database.
func (uc *UserController) saveUser(c *gin.Context) {
	var newUser user.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := uc.UserService.SaveUser(context.Background(), &newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User data saved successfully!"})
}

// fetchUserList handles the GET request to fetch a list of all users from the database.
func (uc *UserController) fetchUserList(c *gin.Context) {
	users, err := uc.UserService.FetchUserList(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// fetchUserById handles the GET request to fetch a specific user by ID from the database.
func (uc *UserController) fetchUserById(c *gin.Context) {
	id := c.Param("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := uc.UserService.FetchUserByID(context.Background(), intID)
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

// deleteUser handles the DELETE request to delete a user by ID from the database.
func (uc *UserController) deleteUser(c *gin.Context) {
	id := c.Param("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := uc.UserService.DeleteUser(context.Background(), intID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user data deleted Successfully"})
}

// updateUser handles the PUT request to update a user by ID with new data.
func (uc *UserController) updateUser(c *gin.Context) {
	id := c.Param("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var updatedUser user.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := uc.UserService.UpdateUser(context.Background(), intID, &updatedUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// getUserNameByName handles the GET request to fetch a user by their name from the database.
func (uc *UserController) getUserNameByName(c *gin.Context) {
	name := c.Param("name")

	user, err := uc.UserService.GetUserNameByName(context.Background(), name)
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
