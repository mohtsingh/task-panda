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

	file, _, err := c.Request().FormFile("photo")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Photo is required"})
	}
	defer file.Close()

	photoBytes, err := io.ReadAll(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to read photo"})
	}

	query := `INSERT INTO profiles (full_name, email, address, phone_number, bio, role, photo)
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var id int
	err = db.DB.QueryRow(query, fullName, email, address, phone, bio, role, photoBytes).Scan(&id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create profile"})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "Profile created",
		"profile": echo.Map{
			"id":       id,
			"fullName": fullName,
			"email":    email,
			"role":     role,
		},
	})
}

func GetProfileByEmail(c echo.Context) error {
	email := c.Param("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Email is required"})
	}

	var profile struct {
		ID          int
		FullName    string
		Email       string
		Address     string
		PhoneNumber string
		Bio         string
		Role        string
		Photo       []byte
	}

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
