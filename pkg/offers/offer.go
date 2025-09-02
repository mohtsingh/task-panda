package offers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"task-panda/pkg/db"

	"github.com/labstack/echo/v4"
)
// UpdateOfferRequest represents the JSON request body for updating an offer
type UpdateOfferRequest struct {
	OfferedPrice *float64 `json:"offered_price,omitempty"`
	Message      *string  `json:"message,omitempty"`
}

// Create an offer for a task
func CreateOffer(c echo.Context) error {
	taskIDStr := c.FormValue("task_id")
	providerIDStr := c.FormValue("provider_id")
	offeredPriceStr := c.FormValue("offered_price")
	message := c.FormValue("message")

	if taskIDStr == "" || providerIDStr == "" || offeredPriceStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "task_id, provider_id, and offered_price are required"})
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid task_id format"})
	}

	providerID, err := strconv.Atoi(providerIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid provider_id format"})
	}

	offeredPrice, err := strconv.ParseFloat(offeredPriceStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid offered_price format"})
	}

	// Check if task exists and is open
	var taskStatus string
	var customerID int
	err = db.DB.QueryRow(`SELECT status, created_by FROM tasks WHERE id = $1`, taskID).Scan(&taskStatus, &customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to check task"})
	}

	if taskStatus != "OPEN" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Task is not open for offers"})
	}

	// Check if provider already made an offer
	var existingOfferID int
	err = db.DB.QueryRow(`SELECT id FROM offers WHERE task_id = $1 AND provider_id = $2`, taskID, providerID).Scan(&existingOfferID)
	if err == nil {
		return c.JSON(http.StatusConflict, echo.Map{"error": "You have already made an offer for this task"})
	}

	// Create the offer
	var offerID int
	var createdAt, updatedAt time.Time
	query := `INSERT INTO offers (task_id, provider_id, offered_price, message) 
	          VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err = db.DB.QueryRow(query, taskID, providerID, offeredPrice, message).Scan(&offerID, &createdAt, &updatedAt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create offer"})
	}

	offer := Offer{
		ID:           offerID,
		TaskID:       taskID,
		ProviderID:   providerID,
		OfferedPrice: offeredPrice,
		Message:      message,
		Status:       "PENDING",
		CreatedAt:    createdAt.Format(time.RFC3339),
		UpdatedAt:    updatedAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusCreated, offer)
}

// Update an existing offer
func UpdateOffer(c echo.Context) error {
	offerIDStr := c.Param("offer_id")
	offerID, err := strconv.Atoi(offerIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid offer_id format"})
	}

	// Parse JSON request body
	var req UpdateOfferRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid JSON format"})
	}

	// Validate that at least one field is provided
	if req.OfferedPrice == nil && req.Message == nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "At least one field (offered_price or message) must be provided"})
	}

	// Get the existing offer to verify ownership and status
	var existingOffer Offer
	query := `SELECT id, task_id, provider_id, offered_price, message, status, created_at, updated_at
	          FROM offers WHERE id = $1`
	err = db.DB.QueryRow(query, offerID).Scan(&existingOffer.ID, &existingOffer.TaskID, 
		&existingOffer.ProviderID, &existingOffer.OfferedPrice, &existingOffer.Message,
		&existingOffer.Status, &existingOffer.CreatedAt, &existingOffer.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Offer not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch offer"})
	}

	// Only allow updates if offer is still pending
	if existingOffer.Status != "PENDING" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Cannot update offer that is not in pending status"})
	}

	// Use existing values if not provided, otherwise use new values
	updatedPrice := existingOffer.OfferedPrice
	updatedMessage := existingOffer.Message

	if req.OfferedPrice != nil {
		updatedPrice = *req.OfferedPrice
	}

	if req.Message != nil {
		updatedMessage = *req.Message
	}

	// Update the offer
	var updatedAt time.Time
	updateQuery := `UPDATE offers SET offered_price = $1, message = $2, updated_at = CURRENT_TIMESTAMP 
	                WHERE id = $3 RETURNING updated_at`
	err = db.DB.QueryRow(updateQuery, updatedPrice, updatedMessage, offerID).Scan(&updatedAt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update offer"})
	}

	// Return updated offer
	updatedOffer := Offer{
		ID:           existingOffer.ID,
		TaskID:       existingOffer.TaskID,
		ProviderID:   existingOffer.ProviderID,
		OfferedPrice: updatedPrice,
		Message:      updatedMessage,
		Status:       existingOffer.Status,
		CreatedAt:    existingOffer.CreatedAt,
		UpdatedAt:    updatedAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, updatedOffer)
}

// Get all offers for a task
func GetTaskOffers(c echo.Context) error {
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid task_id format"})
	}

	query := `SELECT o.id, o.task_id, o.provider_id, o.offered_price, o.message, o.status, 
	          o.created_at, o.updated_at, p.full_name 
	          FROM offers o 
	          JOIN profiles p ON o.provider_id = p.id 
	          WHERE o.task_id = $1 ORDER BY o.created_at ASC`

	rows, err := db.DB.Query(query, taskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch offers"})
	}
	defer rows.Close()

	var offers []Offer
	for rows.Next() {
		var o Offer
		err := rows.Scan(&o.ID, &o.TaskID, &o.ProviderID, &o.OfferedPrice, &o.Message,
			&o.Status, &o.CreatedAt, &o.UpdatedAt, &o.ProviderName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to parse offer data"})
		}
		offers = append(offers, o)
	}

	return c.JSON(http.StatusOK, offers)
}

// Accept an offer
func AcceptOffer(c echo.Context) error {
	offerIDStr := c.Param("offer_id")
	offerID, err := strconv.Atoi(offerIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid offer_id format"})
	}

	// Get offer details
	var offer Offer
	var customerID int
	query := `SELECT o.id, o.task_id, o.provider_id, o.offered_price, o.status, t.created_by
	          FROM offers o 
	          JOIN tasks t ON o.task_id = t.id 
	          WHERE o.id = $1`
	err = db.DB.QueryRow(query, offerID).Scan(&offer.ID, &offer.TaskID, &offer.ProviderID,
		&offer.OfferedPrice, &offer.Status, &customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Offer not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch offer"})
	}

	if offer.Status != "PENDING" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Offer is not in pending status"})
	}

	// Start transaction
	tx, err := db.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to start transaction"})
	}
	defer tx.Rollback()

	// Update task status and set accepted provider
	_, err = tx.Exec(`UPDATE tasks SET status = 'ACCEPTED', accepted_provider_id = $1 
	                  WHERE id = $2`, offer.ProviderID, offer.TaskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update task"})
	}

	// Accept this offer
	_, err = tx.Exec(`UPDATE offers SET status = 'ACCEPTED' WHERE id = $1`, offerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to accept offer"})
	}

	// Reject all other offers for this task
	_, err = tx.Exec(`UPDATE offers SET status = 'REJECTED' 
	                  WHERE task_id = $1 AND id != $2`, offer.TaskID, offerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to reject other offers"})
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message":  "Offer accepted successfully",
		"offer_id": offerID,
		"task_id":  offer.TaskID,
	})
}
