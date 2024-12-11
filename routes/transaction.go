package routes

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/halosatrio/xwing/models"
)

// type transactionQueryReq struct {
// 	DateStart string `json:"email" binding:"required,email"`
// 	DateEnd   string `json:"password" binding:"required,min=8"`
// 	Category  string `json:"category" binding:"required,min=8"`
// }

func GetAllTransactions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// var transactionReq transactionQueryReq
		var transactions models.TransactionSchema
		userID, _ := c.MustGet("user_id").(float64)
		// email, _ := c.MustGet("email").(string)

		queryGetTransaction := `
			SELECT id, user_id, type, amount, category, date, notes, is_active, created_at, updated_at
			FROM swordfish.transactions
			WHERE
				user_id=$1
				AND
				is_active=true
			LIMIT 200	
		`

		err := db.QueryRow(queryGetTransaction, userID).
			Scan(
				&transactions.ID,
				&transactions.UserId,
				&transactions.Type,
				&transactions.Amount,
				&transactions.Category,
				&transactions.Date,
				&transactions.Notes,
				&transactions.IsActive,
				&transactions.CreatedAt,
				&transactions.UpdatedAt,
			)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to fetch transactions!",
				"error":   err.Error(),
			})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success Login!",
			"data":    "hehe",
		})
	}
}
