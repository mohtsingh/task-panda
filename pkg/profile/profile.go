package profile

import (
	"database/sql"
	"net/http"

	"task-panda/pkg/db"

	"github.com/labstack/echo/v4"
)

type CreateProfileRequest struct {
	FullName    string `json:"full_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	Bio         string `json:"bio"`
	Role        string `json:"role" validate:"required"`
}

type UpdateProfileRequest struct {
	FullName    string `json:"full_name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	Bio         string `json:"bio"`
	Role        string `json:"role"`
}

func CreateProfile(c echo.Context) error {
	var req CreateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid JSON payload"})
	}

	if req.FullName == "" || req.Email == "" || req.Role == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing required fields"})
	}

	var existingEmail string
	err := db.DB.QueryRow(`SELECT email FROM profiles WHERE email = $1`, req.Email).Scan(&existingEmail)
	if err == nil {
		return c.JSON(http.StatusConflict, echo.Map{"error": "Email already exists"})
	}

	query := `INSERT INTO profiles (full_name, email, address, phone_number, bio, role)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	var id int
	err = db.DB.QueryRow(query, req.FullName, req.Email, req.Address, req.PhoneNumber, req.Bio, req.Role).Scan(&id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create profile"})
	}

	// Create response using the Profile struct
	profile := Profile{
		ID:          id,
		FullName:    req.FullName,
		Email:       req.Email,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
		Bio:         req.Bio,
		Role:        req.Role,
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "Profile created",
		"profile": profile,
	})
}

func UpdateProfile(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "ID is required"})
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid JSON payload"})
	}

	// Check if profile exists
	var existingProfile Profile
	checkQuery := `SELECT id, full_name, email, address, phone_number, bio, role FROM profiles WHERE id = $1`
	err := db.DB.QueryRow(checkQuery, id).Scan(&existingProfile.ID, &existingProfile.FullName,
		&existingProfile.Email, &existingProfile.Address, &existingProfile.PhoneNumber,
		&existingProfile.Bio, &existingProfile.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Profile not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch profile"})
	}

	// Use existing values if not provided in request
	fullName := req.FullName
	if fullName == "" {
		fullName = existingProfile.FullName
	}

	address := req.Address
	if address == "" {
		address = existingProfile.Address
	}

	phone := req.PhoneNumber
	if phone == "" {
		phone = existingProfile.PhoneNumber
	}

	bio := req.Bio
	if bio == "" {
		bio = existingProfile.Bio
	}

	role := req.Role
	if role == "" {
		role = existingProfile.Role
	}

	// Update the profile
	updateQuery := `UPDATE profiles SET full_name = $1, address = $2, phone_number = $3, bio = $4, role = $5 WHERE id = $6`
	_, err = db.DB.Exec(updateQuery, fullName, address, phone, bio, role, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update profile"})
	}

	// Return updated profile
	updatedProfile := Profile{
		ID:          existingProfile.ID,
		FullName:    fullName,
		Email:       existingProfile.Email,
		Address:     address,
		PhoneNumber: phone,
		Bio:         bio,
		Role:        role,
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Profile updated",
		"profile": updatedProfile,
	})
}

func GetProfileByEmail(c echo.Context) error {
	email := c.Param("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Email is required"})
	}

	var profile Profile

	query := `SELECT id, full_name, email, address, phone_number, bio, role 
	          FROM profiles WHERE email = $1`
	err := db.DB.QueryRow(query, email).Scan(&profile.ID, &profile.FullName, &profile.Email,
		&profile.Address, &profile.PhoneNumber, &profile.Bio, &profile.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Profile not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch profile"})
	}

	return c.JSON(http.StatusOK, profile)
}
