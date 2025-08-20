package tasks

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"task-panda/pkg/db"

	"github.com/labstack/echo/v4"
)

func CreateTask(c echo.Context) error {
	// Parse form data instead of JSON

	category := c.FormValue("category")
	title := c.FormValue("title")
	description := c.FormValue("description")
	budgetStr := c.FormValue("budget")
	location := c.FormValue("location")
	date := c.FormValue("date")
	createdByStr := c.FormValue("created_by")

	// Validate required fields
	if category == "" || title == "" || description == "" || budgetStr == "" || location == "" || date == "" || createdByStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "All fields are required"})
	}

	// Convert budget string to float64
	budget, err := strconv.ParseFloat(budgetStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid budget format"})
	}

	// Convert created_by to int
	createdBy, err := strconv.Atoi(createdByStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid created_by format"})
	}

	// Create task object
	newTask := Task{
		Category:    category,
		Title:       title,
		Description: description,
		Budget:      budget,
		Location:    location,
		Date:        date,
		CreatedBy:   createdBy,
		Status:      "OPEN",
	}

	// Handle optional image upload
	file, header, err := c.Request().FormFile("image")
	var imageData []byte
	if err == nil {
		defer file.Close()
		imageData, err = io.ReadAll(file)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to read image"})
		}
		fmt.Printf("Uploaded image: %s, size: %d bytes\n", header.Filename, len(imageData))
		// You can save imageData to database or file system here
	}

	// Insert task into database
	query := `INSERT INTO tasks (category, title, description, budget, location, date, created_by, status) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	err = db.DB.QueryRow(query, newTask.Category, newTask.Title, newTask.Description, newTask.Budget,
		newTask.Location, newTask.Date, newTask.CreatedBy, newTask.Status).Scan(&newTask.ID, &newTask.CreatedAt, &newTask.UpdatedAt)
	if err != nil {
		fmt.Printf("Database error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to insert task"})
	}

	fmt.Printf("Task created successfully: %+v\n", newTask)
	return c.JSON(http.StatusCreated, newTask)
}
func GetTaskByID(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID format"})
	}

	var task Task
	query := `SELECT id, category, title, description, budget, location, date, created_by, status, 
	          accepted_provider_id, created_at, updated_at FROM tasks WHERE id = $1`
	err = db.DB.QueryRow(query, id).Scan(&task.ID, &task.Category, &task.Title, &task.Description,
		&task.Budget, &task.Location, &task.Date, &task.CreatedBy, &task.Status,
		&task.AcceptedProviderID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch task"})
	}

	return c.JSON(http.StatusOK, task)
}

func GetAllTasks(c echo.Context) error {
	createdBy := c.QueryParam("created_by")

	var rows *sql.Rows
	var err error

	if createdBy != "" {
		rows, err = db.DB.Query(`SELECT id, category, title, description, budget, location, date, 
			created_by, status, accepted_provider_id, created_at, updated_at 
			FROM tasks WHERE created_by = $1 ORDER BY created_at DESC`, createdBy)
	} else {
		// Otherwise fetch all
		rows, err = db.DB.Query(`SELECT id, category, title, description, budget, location, date, 
			created_by, status, accepted_provider_id, created_at, updated_at 
			FROM tasks ORDER BY created_at DESC`)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch tasks"})
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Category, &t.Title, &t.Description, &t.Budget,
			&t.Location, &t.Date, &t.CreatedBy, &t.Status, &t.AcceptedProviderID,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to parse task data"})
		}
		tasks = append(tasks, t)
	}

	return c.JSON(http.StatusOK, tasks)
}

// Update task status (for completing tasks, etc.)
func UpdateTaskStatus(c echo.Context) error {
	taskIDStr := c.Param("task_id")
	status := c.FormValue("status")

	if status == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Status is required"})
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid task_id format"})
	}

	// Validate status
	validStatuses := map[string]bool{
		"OPEN": true, "ACCEPTED": true, "IN_PROGRESS": true,
		"COMPLETED": true, "CANCELLED": true,
	}
	if !validStatuses[status] {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid status"})
	}

	_, err = db.DB.Exec(`UPDATE tasks SET status = $1 WHERE id = $2`, status, taskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update task status"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Task status updated successfully"})
}
