package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
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
	clientURL := os.Getenv("CLIENT_URL")

	r := gin.Default()

	// Custom CORS configuration
	corsConfig := cors.Config{
		// List allowed origins
		AllowOrigins: []string{"http://localhost:3000", clientURL},
		// Allow specific methods
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		// Allow specific headers
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		// Expose headers to the browser
		ExposeHeaders: []string{"X-Custom-Header"},
		// Allow cookies to be sent with the request
		AllowCredentials: true,
		// Cache the preflight response for 12 hours
		MaxAge: 12 * time.Hour,
	}

	// Apply the custom CORS configuration
	r.Use(cors.New(corsConfig))

	// AS BASEPATH
	v1 := r.Group("/v1")
	{
		// health check
		v1.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  http.StatusOK,
				"message": "Welcome to Xwing!",
			})
		})

		// Register and Login [PUBLIC ROUTES]
		v1.POST("/auth/register", routes.RegisterRoute(db))
		v1.POST("/auth/login", routes.LoginUser(db))

		// [PRIVATE ROUTES]
		v1.Use(utils.JWTAuth())

		// auth user
		v1.GET("/auth/user", routes.GetUser())

		// Transaction Routes
		v1.GET("/transaction", routes.GetAllTransactions(db))
		v1.GET("/transaction/:id", routes.GetTransactionById(db))
		v1.POST("/transaction/create", routes.PostCreateTransaction(db))
		v1.PUT("/transaction/:id", routes.PutUpdateTransaction(db))
		v1.DELETE("/transaction/:id", routes.DeleteTransaction(db))
		v1.GET("/transaction/monthly-summary", routes.GetMonthlySummary(db))

		// Report Routes
		v1.GET("/report/quarter/essentials", routes.GetQuarterEssentials(db))
		v1.GET("/report/quarter/non-essentials", routes.GetQuarterNonEssentials(db))
		v1.GET("/report/quarter/shopping", routes.GetQuarterShopping(db))
		v1.GET("/report/annual/cashflow", routes.GetAnnualCashflow(db))
		// GET Annual (WIP, this is for all months per caetgory)
		//.GET("/report/annual", routes.GetAnnualReport(db))

		// Asset Routes
		v1.GET("/asset", routes.GetAsset(db))
		v1.POST("/asset/create", routes.PostCreateAsset(db))

	}
	return r
}
