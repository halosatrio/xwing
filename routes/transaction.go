package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/halosatrio/xwing/models"
)

type transactionQueryReq struct {
	DateStart string `form:"date_start"`
	DateEnd   string `form:"date_end"`
	Category  string `form:"category"`
}

func GetAllTransactions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var transactions []models.TransactionSchema
		var queryReq transactionQueryReq

		// Bind query parameters
		if err := c.BindQuery(&queryReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid query parameters!",
				"errors":  err.Error(),
			})
			return
		}

		userID, ok := c.MustGet("user_id").(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Unauthorized user",
			})
			return
		}

		// Base query
		query := `
			SELECT id, user_id, type, amount, category, date, notes, is_active, created_at, updated_at
			FROM swordfish.transactions
			WHERE user_id=$1 AND is_active=true
		`

		// Query parameters for filtering
		var args []interface{}
		args = append(args, userID)

		if queryReq.DateStart != "" {
			query += " AND date >= $2"
			args = append(args, queryReq.DateStart)
		}
		if queryReq.DateEnd != "" {
			query += " AND date <= $3"
			args = append(args, queryReq.DateEnd)
		}
		if queryReq.Category != "" {
			query += " AND category = $4"
			args = append(args, queryReq.Category)
		}

		query += " ORDER BY date ASC LIMIT 200"

		// Execute query
		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to fetch transactions!",
				"error":   err.Error(),
			})
			return
		}
		defer rows.Close()

		// Scan rows
		for rows.Next() {
			var transaction models.TransactionSchema
			err := rows.Scan(
				&transaction.ID,
				&transaction.UserId,
				&transaction.Type,
				&transaction.Amount,
				&transaction.Category,
				&transaction.Date,
				&transaction.Notes,
				&transaction.IsActive,
				&transaction.CreatedAt,
				&transaction.UpdatedAt,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Failed to parse transaction data!",
					"error":   err.Error(),
				})
				return
			}
			transactions = append(transactions, transaction)
		}

		// Check for row iteration errors
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Error iterating over transactions!",
				"error":   err.Error(),
			})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success!",
			"data":    transactions,
		})
	}
}

type transactionID struct {
	ID string `uri:"id" binding:"required"`
}

func GetTransactionById(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var txID transactionID
		var transaction models.TransactionSchema

		// Validate URI parameter
		if err := c.ShouldBindUri(&txID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid URI parameter!",
				"errors":  err.Error(),
			})
			return
		}

		// Convert ID to integer
		id, err := strconv.Atoi(txID.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Transaction ID must be an integer!",
				"errors":  err.Error(),
			})
			return
		}

		userID, ok := c.MustGet("user_id").(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Unauthorized user",
			})
			return
		}

		// query
		query := `
			SELECT id, user_id, type, amount, category, date, notes, is_active, created_at, updated_at
			FROM swordfish.transactions
			WHERE user_id=$1 AND is_active=true AND id=$2
		`

		err = db.QueryRow(query, userID, id).
			Scan(
				&transaction.ID,
				&transaction.UserId,
				&transaction.Type,
				&transaction.Amount,
				&transaction.Category,
				&transaction.Date,
				&transaction.Notes,
				&transaction.IsActive,
				&transaction.CreatedAt,
				&transaction.UpdatedAt,
			)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Transaction not found",
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

		// success response
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success!",
			"data":    transaction,
		})
	}
}
