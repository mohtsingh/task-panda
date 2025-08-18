package tasks

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"task-panda/pkg/db"

	"github.com/labstack/echo/v4"
)

func CreateTask(c echo.Context) error {
	category := c.FormValue("category")
	title := c.FormValue("title")
	description := c.FormValue("description")
	budgetStr := c.FormValue("budget")
	location := c.FormValue("location")
	date := c.FormValue("date")

	if category == "" || title == "" || description == "" || budgetStr == "" || location == "" || date == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "All fields are required"})
	}

	budget, err := strconv.ParseFloat(budgetStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid budget format"})
	}

	newTask := Task{
		Category:    category,
		Title:       title,
		Description: description,
		Budget:      budget,
		Location:    location,
		Date:        date,
	}

	file, header, err := c.Request().FormFile("image")
	if err == nil {
		defer file.Close()
		imageData, err := io.ReadAll(file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to read image"})
		}
		fmt.Printf("Uploaded image: %s, size: %d bytes\n", header.Filename, len(imageData))
	}

	query := `INSERT INTO tasks (category, title, description, budget, location, date) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = db.DB.QueryRow(query, newTask.Category, newTask.Title, newTask.Description,
		newTask.Budget, newTask.Location, newTask.Date).Scan(&newTask.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to insert task"})
	}

	return c.JSON(http.StatusCreated, newTask)
}

func GetTaskByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID format"})
	}

	var task Task
	query := `SELECT id, category, title, description, budget, location, date FROM tasks WHERE id = $1`
	err = db.DB.QueryRow(query, id).Scan(&task.ID, &task.Category, &task.Title,
		&task.Description, &task.Budget, &task.Location, &task.Date)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Task not found"})
	}

	return c.JSON(http.StatusOK, task)
}

func GetAllTasks(c echo.Context) error {
	rows, err := db.DB.Query(`SELECT id, category, title, description, budget, location, date FROM tasks`)
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
