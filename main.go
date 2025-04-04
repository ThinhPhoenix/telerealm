package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"main.go/handlers"
	"main.go/initializers"
	"main.go/middleware"
	"main.go/repositories"
	"main.go/services"
)

func main() {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	r.Use(cors.New(config))

	h := initializeHandlers()

	// Public endpoints
	r.GET("/ping", h.Ping)
	r.GET("/drive/:key", h.DownloadFile)

	// Protected endpoints
	auth := r.Group("/")
	auth.Use(middleware.AuthRequired())
	{
		auth.POST("/send", h.SendFile)
		auth.GET("/url", h.GetFileURL)
		auth.GET("/info", h.GetFileInfo)
		auth.GET("/verify", h.CheckBotAndChat)
	}

	r.Run(":" + os.Getenv("PORT"))
}

func initializeHandlers() *handlers.Handlers {
	initializers.LoadEnvironment()

	repo := initializeRepositories()
	service := services.NewFileService(repo)
	return handlers.NewHandlers(service)
}

func initializeRepositories() repositories.FileRepository {
	return repositories.NewFileRepository()
}
