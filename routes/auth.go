package routes

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/halosatrio/xwing/models"
	"golang.org/x/crypto/bcrypt"
)

// User represents the structure of the request body for registration
type registerUserReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterRoute handles user registration
func RegisterRoute(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userReq registerUserReq
		var existingUser models.UserSchema

		// Validate request body
		if err := c.ShouldBindJSON(&userReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"message": "Failed to register user!",
				"errors":  err.Error(),
			})
			return
		}

		// Check if email already exists
		queryCheckEmail := `
			SELECT id, username, email
			FROM swordfish.users
			WHERE email = $1
		`
		err := db.QueryRow(queryCheckEmail, userReq.Email).
			Scan(&existingUser.ID, &existingUser.Username, &existingUser.Email)
		if err != nil && err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to check existing email",
				"error":   err.Error(),
			})
			return
		}

		if err == nil {
			// Email already exists
			c.JSON(http.StatusConflict, gin.H{
				"status":  http.StatusConflict,
				"message": "Email is already registered",
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

		var newUser models.UserSchema
		err = db.QueryRow(query, userReq.Username, userReq.Email, string(hashedPassword), time.Now()).
			Scan(&newUser.ID, &newUser.Username, &newUser.Email)
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
		var user models.UserSchema

		// Validate request body
		if err := c.ShouldBindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Failed to login!",
				"errors":  err.Error(),
			})
			return
		}

		// Query user by email
		queryGetUserByEmail := `
			SELECT id, username, email, password
			FROM swordfish.users
			WHERE email=$1
		`
		err := db.QueryRow(queryGetUserByEmail, loginReq.Email).
			Scan(&user.ID, &user.Username, &user.Email, &user.Password)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Invalid email or password",
			})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Database error",
				"error":   err.Error(),
			})
			return
		}

		// Compare password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "invalid credentials",
				"error":   err.Error(),
			})
			return
		}

		// Generate JWT
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Server misconfiguration",
				"error":   "JWT secret is not set",
			})
			return
		}

		// Set token expiration
		tokenDuration := time.Hour * 24 // Default
		if envDuration := os.Getenv("JWT_EXPIRATION_HOURS"); envDuration != "" {
			if hours, err := strconv.Atoi(envDuration); err == nil {
				tokenDuration = time.Duration(hours) * time.Hour
			}
		}

		// Create token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":   user.ID,
			"email": user.Email,
			"exp":   time.Now().Add(tokenDuration).Unix(),
		})

		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to generate token",
				"error":   "[auth][jwt]" + err.Error(),
			})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success Login!",
			"data":    tokenString,
		})
	}
}
