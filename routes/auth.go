package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/halosatrio/xwing/models"
	"golang.org/x/crypto/bcrypt"
)

// User represents the structure of the request body for registration
type registerUserReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

var user = &models.UserSchema{}

// RegisterRoute handles user registration
func RegisterRoute(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userReq registerUserReq

		// Validate request body
		if err := c.ShouldBindJSON(&userReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"message": "Failed to register user!",
				"errors":  err.Error(),
			})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), bcrypt.DefaultCost)
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

		err = db.QueryRow(query, userReq.Username, userReq.Email, string(hashedPassword), time.Now()).
			Scan(&user.ID, &user.Username, &user.Email)
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
			"status":  http.StatusOK,
			"message": "User registered successfully!",
		})
	}
}

type loginUserReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func LoginUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginReq loginUserReq

		if err := c.ShouldBindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"message": "Failed to login!",
				"errors":  err.Error(),
			})
			return
		}

		queryGetUserByEmail := `
			SELECT id, username, email, password
			FROM swordfish.users
			WHERE email=$1
		`

		err := db.QueryRow(queryGetUserByEmail, loginReq.Email).
			Scan(&user.ID, &user.Username, &user.Email, &user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  500,
				"message": "Failed to get user by email",
				"error":   err.Error(),
			})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  500,
				"message": "invalid credentials",
				"error":   err.Error(),
			})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"message": "Success Login!",
		})
	}
}
