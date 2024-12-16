package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/halosatrio/xwing/models"
	"github.com/halosatrio/xwing/utils"
)

func GetAsset(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var assets []models.AssetSchema

		// get userid jwt
		userID, _ := c.MustGet("user_id").(float64)

		query := `
			SELECT id, user_id, account, amount, date, COALESCE(notes, '') as notes, created_at, updated_at
			FROM swordfish.assets
			WHERE user_id = $1
			LIMIT 200
		`

		rows, err := db.Query(query, userID)
		if err != nil {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch transactions!", err.Error())
			return
		}
		defer rows.Close()

		for rows.Next() {
			var asset models.AssetSchema
			err := rows.Scan(
				&asset.ID,
				&asset.UserId,
				&asset.Account,
				&asset.Amount,
				&asset.Date,
				&asset.Notes,
				&asset.CreatedAt,
				&asset.UpdatedAt,
			)
			if err != nil {
				utils.RespondError(c, http.StatusInternalServerError, "Failed to parse transaction data!", err.Error())
				return
			}
			assets = append(assets, asset)
		}

		// success
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success!",
			"data":    assets,
		})
	}
}

type createAssetReq struct {
	Account string `form:"account" binding:"required"`
	Amount  int    `form:"amount" binding:"required"`
	Date    string `form:"date" binding:"required"`
	Notes   string `form:"notes"`
}

func PostCreateAsset(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var assetReq createAssetReq

		// Validate request body
		if err := c.ShouldBindJSON(&assetReq); err != nil {
			utils.RespondError(c, http.StatusBadRequest, "Failed to create asset!", err.Error())
			return
		}

		// get userid jwt
		userID, _ := c.MustGet("user_id").(float64)
		query := `
			INSERT INTO swordfish.assets (user_id, account, amount, date, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, user_id, account, amount, date, notes, created_at, updated_at
		`

		var newAsset models.AssetSchema
		err := db.QueryRow(query, userID, assetReq.Account, assetReq.Amount, assetReq.Date, assetReq.Notes, time.Now(), time.Now()).
			Scan(&newAsset.ID, &newAsset.UserId, &newAsset.Account, &newAsset.Amount, &newAsset.Date, &newAsset.Notes, &newAsset.CreatedAt, &newAsset.UpdatedAt)
		if err != nil {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to insert asset into database!", err.Error())
			return
		}

		// success respond
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Success!",
			"data":    newAsset,
		})
	}
}
