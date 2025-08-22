package routes

import (
	"task-panda/pkg/profile"
	"task-panda/pkg/tasks"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	// Task routes
	e.POST("/tasks", tasks.CreateTask)
	e.GET("/tasks/:id", tasks.GetTaskByID)
	e.GET("/tasks", tasks.GetAllTasks)

	// Profile routes
	e.POST("/profile", profile.CreateProfile)
	e.PUT("/profile/:id", profile.UpdateProfile)
	e.GET("/profile/:email", profile.GetProfileByEmail)
}
