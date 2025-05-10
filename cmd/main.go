package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
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

func main() {
	initDB()
	defer db.Close()

	e := echo.New()
	e.POST("/tasks", createTask)
	e.GET("/tasks/:id", getTaskByID)

	e.Logger.Fatal(e.Start(":8080"))
}
