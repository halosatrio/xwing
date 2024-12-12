package routes

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

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

type transactionReq struct {
	Type     string `json:"type" binding:"required"`
	Amount   int    `json:"amount" binding:"required"`
	Category string `json:"category" binding:"required"`
	Date     string `json:"date" binding:"required"`
	Notes    string `json:"notes"`
}

func PostCreateTransaction(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var createTxReq transactionReq

		// Validate request body
		if err := c.ShouldBindJSON(&createTxReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"message": "Failed to create transaction!",
				"errors":  err.Error(),
			})
			return
		}

		// Validate user from JWT
		userID, ok := c.MustGet("user_id").(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Unauthorized user",
			})
			return
		}

		query := `
      INSERT INTO swordfish.transactions ( user_id, type, amount, category, date, notes, created_at, updated_at)
      VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
      RETURNING id, user_id, type, amount, category, date, notes, created_at, updated_at
    `
		var newTransaction models.TransactionSchema
		err := db.QueryRow(query, userID, createTxReq.Type, createTxReq.Amount, createTxReq.Category, createTxReq.Date, createTxReq.Notes, time.Now(), time.Now()).
			Scan(
				&newTransaction.ID,
				&newTransaction.UserId,
				&newTransaction.Type,
				&newTransaction.Amount,
				&newTransaction.Category,
				&newTransaction.Date,
				&newTransaction.Notes,
				&newTransaction.CreatedAt,
				&newTransaction.UpdatedAt,
			)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  500,
				"message": "Failed to insert transaction into database!",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success create transaction!",
			"data":    newTransaction,
		})
	}
}

func PutUpdateTransaction(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var txID transactionID
		var updateTxReq transactionReq

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

		// Validate request body
		if err := c.ShouldBindJSON(&updateTxReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"message": "Failed to update transaction!",
				"errors":  err.Error(),
			})
			return
		}

		var updatedTransaction models.TransactionSchema
		query := `
			UPDATE swordfish.transactions
			SET type = $1, amount = $2, category = $3, date = $4, notes = $5, updated_at = $6
			WHERE id = $7 AND user_id = $8 AND is_active = true
			RETURNING id, user_id, type, amount, category, date, notes, is_active, created_at, updated_at
		`
		err = db.QueryRow(query, updateTxReq.Type, updateTxReq.Amount, updateTxReq.Category, updateTxReq.Date, updateTxReq.Notes, time.Now(), id, userID).
			Scan(
				&updatedTransaction.ID,
				&updatedTransaction.UserId,
				&updatedTransaction.Type,
				&updatedTransaction.Amount,
				&updatedTransaction.Category,
				&updatedTransaction.Date,
				&updatedTransaction.Notes,
				&updatedTransaction.IsActive,
				&updatedTransaction.CreatedAt,
				&updatedTransaction.UpdatedAt,
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

		// return status success
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success update transaction!",
			"data":    updatedTransaction,
		})
	}
}

func DeleteTransaction(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var txID transactionID

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

		query := `
      UPDATE swordfish.transactions
      SET is_active = false, updated_at = $1
      WHERE id = $2 AND user_id = $3 AND is_active = true
    `
		result, err := db.Exec(query, time.Now(), id, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Database error",
				"error":   err.Error(),
			})
			return
		}

		// Check if any rows were affected
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Error checking update result",
				"error":   err.Error(),
			})
			return
		}

		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Transaction not found or already inactive",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Transaction deleted successfully",
		})
	}
}

type monthlySummaryQueryReq struct {
	DateStart string `form:"date_start" binding:"required"`
	DateEnd   string `form:"date_end" binding:"required"`
}

type monthlySummaryData struct {
	Category    string `json:"category"`
	TotalAmount int    `json:"total_amount"`
	Count       int    `json:"count"`
}

func GetMonthlySummary(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var queryReq monthlySummaryQueryReq
		var summaryData []monthlySummaryData

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

		summaryQuery := `
      SELECT category, SUM(amount) AS total_amount, COUNT(id) as count
      FROM swordfish.transactions as tx
      WHERE tx.user_id = $1 AND tx.is_active = true AND tx.date BETWEEN $2 AND $3
      GROUP BY category;
    `
		// Execute query
		rows, err := db.Query(summaryQuery, userID, queryReq.DateStart, queryReq.DateEnd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to fetch summary transaction!",
				"error":   err.Error(),
			})
			return
		}
		defer rows.Close()

		// Scan rows
		for rows.Next() {
			var summary monthlySummaryData
			err := rows.Scan(
				&summary.Category,
				&summary.TotalAmount,
				&summary.Count,
			)
			log.Println(summary)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  http.StatusInternalServerError,
					"message": "Failed to parse summary data!",
					"error":   err.Error(),
				})
				return
			}
			summaryData = append(summaryData, summary)
		}
		// Check for row iteration errors
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Error iterating over summary transactions!",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Successs!",
			"data":    summaryData,
		})
	}
}
