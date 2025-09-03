package notifications

import (
	"database/sql"
	"log"
	"net/http"
	"task-panda/pkg/db"
	"time"

	"github.com/labstack/echo/v4"
)

func NotifyServiceProviders(taskID int) {
	query := `
        SELECT p.id, dt.token 
        FROM profiles p 
        INNER JOIN device_tokens dt ON p.id = dt.profile_id 
        WHERE p.role = 'SERVICE_PROVIDER' AND dt.is_active = true`

	rows, err := db.DB.Query(query)
	if err != nil {
		log.Printf("Failed to fetch service providers and tokens: %v\n", err)
		return
	}
	defer rows.Close()

	// Use map to count notifications per profile
	profileNotifications := make(map[int]int)
	totalNotifications := 0

	for rows.Next() {
		var profileID int
		var token string

		err = rows.Scan(&profileID, &token)
		if err != nil {
			log.Printf("Failed to scan row: %v\n", err)
			continue
		}

		// Mock notification sending
		log.Printf("Mocking Notification for task id %d, profile id %d\n", taskID, profileID)
		profileNotifications[profileID]++
		totalNotifications++
	}

	// Log summary
	for profileID, count := range profileNotifications {
		log.Printf("Sent %d notifications to profile %d\n", count, profileID)
	}

	log.Printf("Notification process completed. Total notifications sent: %d\n", totalNotifications)
}

func RegisterDeviceToken(c echo.Context) error {
	var req RegisterTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid JSON format"})
	}

	// Validate required fields
	if req.ProfileID == 0 || req.Token == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "profile_id, and token are required",
		})
	}

	// Check if profile exists
	var profileExists bool
	err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM profiles WHERE id = $1)`, req.ProfileID).Scan(&profileExists)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to verify profile"})
	}
	if !profileExists {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Profile not found"})
	}

	// Check if token already exists for this profile and platform
	var existingTokenID int
	err = db.DB.QueryRow(`SELECT id FROM device_tokens WHERE profile_id = $1 AND is_active = true`,
		req.ProfileID).Scan(&existingTokenID)

	if err == sql.ErrNoRows {
		// Create new token
		var tokenID int
		var createdAt time.Time
		query := `INSERT INTO device_tokens (profile_id, token, platform, is_active) 
		          VALUES ($1, $2, $3, true) RETURNING id, created_at`
		err = db.DB.QueryRow(query, req.ProfileID, req.Token, req.Platform).Scan(&tokenID, &createdAt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to register token"})
		}

		return c.JSON(http.StatusCreated, echo.Map{
			"message": "Device token registered successfully",
		})
	} else if err == nil {
		// Update existing token
		_, err = db.DB.Exec(`UPDATE device_tokens SET token = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
			req.Token, existingTokenID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update token"})
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "Device token updated successfully",
		})
	} else {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Database error"})
	}
}
