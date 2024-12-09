package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// User represents the structure of the request body for registration
type User struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterRoute handles user registration
func RegisterRoute(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user User

		// Validate request body
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"message": "Failed to register user!",
				"errors":  err.Error(),
			})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  500,
				"message": "Failed to hash password!",
				"error":   err.Error(),
			})
			return
		}

		// Insert user into the database
		query := `
			INSERT INTO swordfish.users (username, email, password, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id, username, email
		`
		var createdUser struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}

		err = db.QueryRow(query, user.Username, user.Email, string(hashedPassword), time.Now()).
			Scan(&createdUser.ID, &createdUser.Username, &createdUser.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  500,
				"message": "Failed to insert user into database!",
				"error":   err.Error(),
			})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"message": "User registered successfully!",
			"data":    createdUser,
		})
	}
}
