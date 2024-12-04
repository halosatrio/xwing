package handlers

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// AS BASEPATH
	v := r.Group("/v1")

	// the routers
	v.GET("/test", Welcome)

	return r
}
