package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

type Task struct {
	ID          int     `json:"id"`
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Budget      float64 `json:"budget"`
	Location    string  `json:"location"`
	Date        string  `json:"date"`
}

type Profile struct {
	ID          int    `json:"id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	Bio         string `json:"bio"`
	Role        string `json:"role"` // "CUSTOMER" or "SERVICE_PROVIDER"
	Photo       []byte `json:"-"`    // Exclude from JSON output
}

var db *sql.DB

func initDB() {
	var err error
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to the database!")
}

func createTask(c echo.Context) error {
	// Parse form data instead of JSON
	category := c.FormValue("category")
	title := c.FormValue("title")
	description := c.FormValue("description")
	budgetStr := c.FormValue("budget")
	location := c.FormValue("location")
	date := c.FormValue("date")

	// Validate required fields
	if category == "" || title == "" || description == "" || budgetStr == "" || location == "" || date == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "All fields are required"})
	}

	// Convert budget string to float64
	budget, err := strconv.ParseFloat(budgetStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid budget format"})
	}

	// Create task object
	newTask := Task{
		Category:    category,
		Title:       title,
		Description: description,
		Budget:      budget,
		Location:    location,
		Date:        date,
	}

	// Handle optional image upload
	file, header, err := c.Request().FormFile("image")
	var imageData []byte
	if err == nil {
		defer file.Close()
		imageData, err = ioutil.ReadAll(file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to read image"})
		}
		fmt.Printf("Uploaded image: %s, size: %d bytes\n", header.Filename, len(imageData))
		// You can save imageData to database or file system here
	}

	// Insert task into database
	query := `INSERT INTO tasks (category, title, description, budget, location, date) 
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = db.QueryRow(query, newTask.Category, newTask.Title, newTask.Description, newTask.Budget, newTask.Location, newTask.Date).Scan(&newTask.ID)
	if err != nil {
		fmt.Printf("Database error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to insert task"})
	}

	fmt.Printf("Task created successfully: %+v\n", newTask)
	return c.JSON(http.StatusCreated, newTask)
}

func getTaskByID(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID format"})
	}

	var item Task
	query := `SELECT id, category, title, description, budget, location, date FROM tasks WHERE id = $1`
	err = db.QueryRow(query, id).Scan(&item.ID, &item.Category, &item.Title, &item.Description, &item.Budget, &item.Location, &item.Date)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch item"})
	}

	return c.JSON(http.StatusOK, item)
}

func getAllTasks(c echo.Context) error {
	rows, err := db.Query(`SELECT id, category, title, description, budget, location, date FROM tasks`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch tasks"})
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Category, &t.Title, &t.Description, &t.Budget, &t.Location, &t.Date); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to parse task data"})
		}
		tasks = append(tasks, t)
	}

	return c.JSON(http.StatusOK, tasks)
}

// Create a new profile with photo upload
func createProfile(c echo.Context) error {
	// Parse form data (including file)
	fullName := c.FormValue("full_name")
	email := c.FormValue("email")
	address := c.FormValue("address")
	phone := c.FormValue("phone_number")
	bio := c.FormValue("bio")
	role := c.FormValue("role")

	if fullName == "" || email == "" || role == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing required fields"})
	}

	// Check if email already exists
	var existingEmail string
	err := db.QueryRow(`SELECT email FROM profiles WHERE email = $1`, email).Scan(&existingEmail)
	if err == nil {
		return c.JSON(http.StatusConflict, echo.Map{"error": "Email already exists"})
	}

	// Retrieve the photo file from the request
	file, _, err := c.Request().FormFile("photo")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Photo is required"})
	}
	defer file.Close()

	// Read the photo file into a byte slice (binary data)
	photoBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to read photo"})
	}

	// Insert profile and photo into the database
	query := `INSERT INTO profiles (full_name, email, address, phone_number, bio, role, photo)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var id int
	err = db.QueryRow(query, fullName, email, address, phone, bio, role, photoBytes).Scan(&id)
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

// Get profile by email
func getProfileByEmail(c echo.Context) error {
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
	err := db.QueryRow(query, email).Scan(&profile.ID, &profile.FullName, &profile.Email, &profile.Address,
		&profile.PhoneNumber, &profile.Bio, &profile.Role, &profile.Photo)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Profile not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch profile"})
	}

	return c.JSON(http.StatusOK, profile)
}

func main() {
	initDB()
	defer db.Close()

	e := echo.New()

	// Enable CORS with more specific settings
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},                                                                // Specify allowed origins
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}, // Allowed methods
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization},                   // Allowed headers
		AllowCredentials: true,                                                                         // Allow credentials (cookies, etc.)
	}))

	e.POST("/tasks", createTask)
	e.GET("/tasks/:id", getTaskByID)
	e.GET("/tasks", getAllTasks)
	e.POST("/profile", createProfile)
	e.GET("/profile/:email", getProfileByEmail)
	e.Logger.Fatal(e.Start(":8080"))
}
