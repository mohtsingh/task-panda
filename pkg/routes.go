package routes

import (
	"task-panda/pkg/notifications"
	"task-panda/pkg/offers"
	"task-panda/pkg/profile"
	"task-panda/pkg/tasks"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	// Task routes
	e.POST("/tasks", tasks.CreateTask)
	e.GET("/tasks/:id", tasks.GetTaskByID)
	e.GET("/tasks", tasks.GetAllTasks)
	e.PUT("/tasks/:task_id/status", tasks.UpdateTaskStatus)

	// Profile routes
	e.POST("/profile", profile.CreateProfile)
	e.GET("/profile/:email", profile.GetProfileByEmail)

	// Offer routes
	e.POST("/offers", offers.CreateOffer)
	e.GET("/tasks/:task_id/offers", offers.GetTaskOffers)
	e.POST("/offers/:offer_id/accept", offers.AcceptOffer)
	e.PUT("/offers/:offer_id", offers.UpdateOffer)
	// Notification routes
	e.POST("/notifications/fcm/token", notifications.RegisterDeviceToken)
}
