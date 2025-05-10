package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
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

var db *sql.DB

func initDB() {
	var err error
	connStr := "postgres://gls:gls@localhost:5433/test?sslmode=disable" // Replace with your credentials
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
	var newTask Task
	if err := c.Bind(&newTask); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	query := `INSERT INTO tasks (category, title, description, budget, location, date) 
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := db.QueryRow(query, newTask.Category, newTask.Title, newTask.Description, newTask.Budget, newTask.Location, newTask.Date).Scan(&newTask.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to insert item"})
	}

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
	rows, err := db.Query("SELECT id, category, title, description, budget, location, date FROM tasks")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch tasks"})
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Category, &task.Title, &task.Description, &task.Budget, &task.Location, &task.Date)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error scanning task"})
		}
		tasks = append(tasks, task)
	}

	return c.JSON(http.StatusOK, tasks)
}

func main() {
	initDB()
	defer db.Close()

	e := echo.New()
	e.POST("/tasks", createTask)
	e.GET("/tasks/:id", getTaskByID)
	e.GET("/tasks", getAllTasks) // GET all tasks

	e.Logger.Fatal(e.Start(":8089"))
}
