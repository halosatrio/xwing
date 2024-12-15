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
		v1.POST("/auth/login", routes.LoginUser(db))
		v1.Use(utils.JWTAuth()).GET("/auth/user", routes.GetUser())

		// Transaction Routes
		v1.Use(utils.JWTAuth()).GET("/transaction", routes.GetAllTransactions(db))
		v1.Use(utils.JWTAuth()).GET("/transaction/:id", routes.GetTransactionById(db))
		v1.Use(utils.JWTAuth()).POST("/transaction/create", routes.PostCreateTransaction(db))
		v1.Use(utils.JWTAuth()).PUT("/transaction/:id", routes.PutUpdateTransaction(db))
		v1.Use(utils.JWTAuth()).DELETE("/transaction/:id", routes.DeleteTransaction(db))
		v1.Use(utils.JWTAuth()).GET("/transaction/monthly-summary", routes.GetMonthlySummary(db))

		// Report Routes
		v1.Use(utils.JWTAuth()).GET("/report/quarter/essentials", routes.GetQuarterEssentials(db))
		v1.Use(utils.JWTAuth()).GET("/report/quarter/non-essentials", routes.GetQuarterNonEssentials(db))
		v1.Use(utils.JWTAuth()).GET("/report/quarter/shopping", routes.GetQuarterShopping(db))
		v1.Use(utils.JWTAuth()).GET("/report/annual", routes.GetAnnualReport(db))
		// GET Annual
		// GET Annual Cashflow

		// Asset Routes
		// GET Assets
		// POST create asset

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
