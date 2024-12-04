package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/halosatrio/xwing/config"
	"github.com/halosatrio/xwing/routes"
)

// main function
func main() {
	config.LoadEnv()

	db := config.ConnectDB()
	defer db.Close()

	r := setupRouter(db)
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
