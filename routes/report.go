package routes

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetReportQuarterEssentials(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success",
		})
	}
}
