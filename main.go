package main

import (
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
	r.GET("/", func(c *gin.Context) {
		c.File("postman_collection.json")
	})

	// Protected endpoints
	auth := r.Group("/")
	auth.Use(middleware.AuthRequired())
	{
		auth.POST("/send", h.SendFile)
		auth.GET("/url", h.GetFileURL)
		auth.GET("/info", h.GetFileInfo)
		auth.GET("/verify", h.CheckBotAndChat)
	}

	r.Run(":6777")
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
