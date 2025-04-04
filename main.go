package main

import (
	"os"

	"telerealm/handlers"
	"telerealm/initializers"
	"telerealm/middleware"
	"telerealm/repositories"
	"telerealm/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	r.Run(":" + func() string {
		p := os.Getenv("PORT")
		if p == "" {
			return "8080"
		}
		return p
	}())
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
