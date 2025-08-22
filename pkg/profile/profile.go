package profile

import (
	"database/sql"
	"io"
	"net/http"

	"task-panda/pkg/db"

	"github.com/labstack/echo/v4"
)

func CreateProfile(c echo.Context) error {
	fullName := c.FormValue("full_name")
	email := c.FormValue("email")
	address := c.FormValue("address")
	phone := c.FormValue("phone_number")
	bio := c.FormValue("bio")
	role := c.FormValue("role")

	if fullName == "" || email == "" || role == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing required fields"})
	}

	var existingEmail string
	err := db.DB.QueryRow(`SELECT email FROM profiles WHERE email = $1`, email).Scan(&existingEmail)
	if err == nil {
		return c.JSON(http.StatusConflict, echo.Map{"error": "Email already exists"})
	}

	// Handle optional photo upload
	var photoBytes []byte
	file, _, err := c.Request().FormFile("photo")
	if err == nil {
		// Photo was provided
		defer file.Close()
		photoBytes, err = io.ReadAll(file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to read photo"})
		}
	}
	// If err != nil, photo wasn't provided, so photoBytes remains nil

	query := `INSERT INTO profiles (full_name, email, address, phone_number, bio, role, photo)
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var id int
	err = db.DB.QueryRow(query, fullName, email, address, phone, bio, role, photoBytes).Scan(&id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create profile"})
	}

	// Create response using the Profile struct
	profile := Profile{
		ID:          id,
		FullName:    fullName,
		Email:       email,
		Address:     address,
		PhoneNumber: phone,
		Bio:         bio,
		Role:        role,
		// Photo is omitted from JSON response due to json:"-" tag
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

	// Check if profile exists
	var existingProfile Profile
	checkQuery := `SELECT id, full_name, email, address, phone_number, bio, role, photo FROM profiles WHERE id = $1`
	err := db.DB.QueryRow(checkQuery, id).Scan(&existingProfile.ID, &existingProfile.FullName,
		&existingProfile.Email, &existingProfile.Address, &existingProfile.PhoneNumber,
		&existingProfile.Bio, &existingProfile.Role, &existingProfile.Photo)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Profile not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch profile"})
	}

	// Get form values (use existing values if not provided)
	fullName := c.FormValue("full_name")
	if fullName == "" {
		fullName = existingProfile.FullName
	}

	address := c.FormValue("address")
	if address == "" {
		address = existingProfile.Address
	}

	phone := c.FormValue("phone_number")
	if phone == "" {
		phone = existingProfile.PhoneNumber
	}

	bio := c.FormValue("bio")
	if bio == "" {
		bio = existingProfile.Bio
	}

	// Handle optional photo upload
	photoBytes := existingProfile.Photo // Keep existing photo by default
	file, _, err := c.Request().FormFile("photo")
	if err == nil {
		// New photo was provided
		defer file.Close()
		photoBytes, err = io.ReadAll(file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to read photo"})
		}
	}

	// Update the profile
	updateQuery := `UPDATE profiles SET full_name = $1, address = $2, phone_number = $3, bio = $4, photo = $5 WHERE id = $6`
	_, err = db.DB.Exec(updateQuery, fullName, address, phone, bio, photoBytes, id)
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
		Role:        existingProfile.Role,
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

	query := `SELECT id, full_name, email, address, phone_number, bio, role, photo 
	          FROM profiles WHERE email = $1`
	err := db.DB.QueryRow(query, email).Scan(&profile.ID, &profile.FullName, &profile.Email,
		&profile.Address, &profile.PhoneNumber, &profile.Bio, &profile.Role, &profile.Photo)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Profile not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch profile"})
	}

	return c.JSON(http.StatusOK, profile)
}
