package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type quarterQueryReq struct {
	Year string `form:"year" binding:"required"`
	Q    string `form:"q" binding:"required"`
}

type Transaction struct {
	Category string `json:"category"`
	Amount   int    `json:"amount"`
}

// QUARTER_MONTH is a mapping for the quarters and months
var QUARTER_MONTH = map[string][]int{
	"1": {1, 2, 3},
	"2": {4, 5, 6},
	"3": {7, 8, 9},
	"4": {10, 11, 12},
}

// getFirstDate returns the first date of a specified year, quarter, and month index.
func getFirstDate(year, q string, month int) (string, error) {
	m, ok := QUARTER_MONTH[q]
	if !ok || month >= len(m) {
		return "", fmt.Errorf("invalid quarter or month")
	}
	date := fmt.Sprintf("%s-%02d-01", year, m[month])
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("invalid date: %v", err)
	}
	return parsed.Format("2006-01-02"), nil
}

// getLastDate returns the last date of a specified year, quarter, and month index.
func getLastDate(year, q string, month int) (string, error) {
	m, ok := QUARTER_MONTH[q]
	if !ok || month >= len(m) {
		return "", fmt.Errorf("invalid quarter or month")
	}
	date := fmt.Sprintf("%s-%02d-01", year, m[month])
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("invalid date: %v", err)
	}
	endOfMonth := parsed.AddDate(0, 1, -1)
	return endOfMonth.Format("2006-01-02"), nil
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

func getQuarterQuery(db *sql.DB, userID float64, date1, date2 string, categories []string) ([]Transaction, error) {
	quoted := make([]string, len(categories))
	for i, s := range categories {
		quoted[i] = fmt.Sprintf("'%s'", s)
	}

	query := `
		SELECT category, SUM(amount) as amount
		FROM swordfish.transactions as tx
		WHERE tx.user_id = $1 
			AND tx.is_active = true 
			AND tx.date BETWEEN $2 AND $3
			AND tx.category IN ($4)
		GROUP BY category
	`
	rows, err := db.Query(query, userID, date1, date2, fmt.Sprintf("(%s)", strings.Join(quoted, ", ")))
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
	// log.Print(rows)
	return result, nil
}

func GetQuarterEssentials(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var queryReq quarterQueryReq

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
		userID, ok := c.MustGet("user_id").(float64)
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

func GetQuarterNonEssentials(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var queryReq quarterQueryReq

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
		userID, ok := c.MustGet("user_id").(float64)
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

		nonEssentials := []string{"misc", "family", "transport", "traveling", "healthcare", "date"}

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
			res, err := getQuarterQuery(db, userID, month[0], month[1], nonEssentials)
			if err != nil {
				log.Printf("Error fetching query: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "Error fetching data"})
				return
			}
			results = append(results, checkCategory(res, nonEssentials))
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
