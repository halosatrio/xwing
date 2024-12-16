package routes

import (
	"database/sql"
	"net/http"

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
