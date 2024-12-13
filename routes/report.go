package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type quarterEssentialsQueryReq struct {
	Year string `form:"year" binding:"required"`
	Q    string `form:"q" binding:"required, oneof=1 2 3 4"`
}

func GetReportQuarterEssentials(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var queryReq quarterEssentialsQueryReq

		// Bind query parameters
		if err := c.BindQuery(&queryReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid query parameters!",
				"errors":  err.Error(),
			})
			return
		}

		// query := `
		// 	SELECT category, amount
		// `

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success",
		})
	}
}

type Transaction struct {
	Category string `json:"category"`
	Amount   int    `json:"amount"`
}

// QUARTER_MONTH maps quarter numbers to their respective months.
var QUARTER_MONTH = map[string][]string{
	"1": {"January", "February", "March"},
	"2": {"April", "May", "June"},
	"3": {"July", "August", "September"},
	"4": {"October", "November", "December"},
}

// getFirstDate returns the first date of a specified year, quarter, and month index.
func getFirstDate(year, quarter string, month int) (string, error) {
	months, ok := QUARTER_MONTH[quarter]
	if !ok || month >= len(months) {
		return "", fmt.Errorf("invalid quarter or month")
	}

	dateStr := fmt.Sprintf("%s %s 01", months[month], year) // Format: "January 2024 01"
	t, err := time.Parse("January 2006 02", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date: %v", err)
	}

	return t.Format("2006-01-02"), nil
}

// getLastDate returns the last date of a specified year, quarter, and month index.
func getLastDate(year, quarter string, month int) (string, error) {
	firstDateStr, err := getFirstDate(year, quarter, month)
	if err != nil {
		return "", err
	}

	t, err := time.Parse("2006-01-02", firstDateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date: %v", err)
	}

	lastDay := t.AddDate(0, 1, -1) // Move to the last day of the month.
	return lastDay.Format("2006-01-02"), nil
}

func checkCategory(resQuery []Transaction, categories []string) []Transaction {
	if len(resQuery) == 0 {
		var result []Transaction
		for _, category := range categories {
			result = append(result, Transaction{Category: category, Amount: 0})
		}
		return result
	}
	reportCategories := make(map[string]bool)
	for _, item := range resQuery {
		reportCategories[item.Category] = true
	}
	var missingItems []Transaction
	for _, category := range categories {
		if !reportCategories[category] {
			missingItems = append(missingItems, Transaction{Category: category, Amount: 0})
		}
	}
	return append(resQuery, missingItems...)
}

func getQuarterQuery(db *sql.DB, userID int, date1, date2 string, categories []string) ([]Transaction, error) {
	query := `
        SELECT category, CAST(SUM(amount) AS INTEGER) as amount
        FROM transactions
        WHERE user_id = $1 AND is_active = true AND date BETWEEN $2 AND $3 AND category = ANY($4)
        GROUP BY category
    `
	rows, err := db.Query(query, userID, date1, date2, categories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Transaction
	for rows.Next() {
		var transaction Transaction
		if err := rows.Scan(&transaction.Category, &transaction.Amount); err != nil {
			return nil, err
		}
		result = append(result, transaction)
	}
	return result, nil
}

func QuarterEssentialsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var queryReq quarterEssentialsQueryReq

		// Bind query parameters
		if err := c.BindQuery(&queryReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid query parameters!",
				"errors":  err.Error(),
			})
			return
		}

		// userID, ok := c.MustGet("user_id").(float64)
		userID, ok := c.MustGet("user_id").(int)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Unauthorized user",
			})
			return
		}

		if _, err := strconv.Atoi(queryReq.Year); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "Year must be a number"})
			return
		}
		if _, err := strconv.Atoi(queryReq.Q); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "Quarter must be a number"})
			return
		}

		essentials := []string{"makan", "cafe", "utils", "errand", "bensin", "olahraga"}

		// Define date ranges for the quarter
		months := [][]string{}
		for i := 0; i < 3; i++ {
			start, err := getFirstDate(queryReq.Year, queryReq.Q, i)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": err.Error()})
				return
			}
			end, err := getLastDate(queryReq.Year, queryReq.Q, i)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": err.Error()})
				return
			}
			months = append(months, []string{start, end})
		}

		var results [][]Transaction
		for _, month := range months {
			res, err := getQuarterQuery(db, userID, month[0], month[1], essentials)
			if err != nil {
				log.Printf("Error fetching query: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "Error fetching data"})
				return
			}
			results = append(results, checkCategory(res, essentials))
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"message": "Success!",
			"data": gin.H{
				"month1": results[0],
				"month2": results[1],
				"month3": results[2],
			},
		})
	}
}
