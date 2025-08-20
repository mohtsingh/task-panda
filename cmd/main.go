package main

import (
	"log"
	"task-panda/pkg/db"
	"task-panda/pkg/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	db.InitDB()
	defer db.DB.Close()
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))
	//FIXME : use the api key when integrated in frontend
	//e.Use(routes.APIKeyMiddleware)
	routes.RegisterRoutes(e)

	log.Fatal(e.Start(":8080"))
}
