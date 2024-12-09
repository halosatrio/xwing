package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/halosatrio/xwing/db"
	"github.com/halosatrio/xwing/routes"
	"github.com/halosatrio/xwing/utils"
	"github.com/joho/godotenv"
)

// main function
func main() {
	// Load Env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// connect DB
	dbx := db.ConnectDB()
	defer dbx.Close()

	// setup routes
	r := setupRouter(dbx)
	r.Run(":8080")
}

// setup app, define routes
func setupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// AS BASEPATH
	v1 := r.Group("/v1")
	{
		// Register Routes
		v1.POST("/auth/register", routes.RegisterRoute(db))
		v1.GET("/auth/login", routes.LoginUser(db))
		v1.Use(utils.JWTAuth()).GET("/auth/user", routes.GetUser())

		// Transaction Routes
		v1.Use(utils.JWTAuth()).GET("/transaction", routes.GetAllTransactions(db))
		v1.Use(utils.JWTAuth()).GET("/transaction/:id", routes.GetTransactionById(db))

		// health check
		v1.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  http.StatusOK,
				"message": "Welcome to Xwing!",
			})
		})
	}

	return r
}
