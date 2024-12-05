package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/halosatrio/xwing/db"
	"github.com/halosatrio/xwing/routes"
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
	v := r.Group("/v1")

	// the routers
	v.GET("/test", routes.Welcome)

	return r
}
