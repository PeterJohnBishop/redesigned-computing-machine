package ginserver

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/subosito/gotenv"
)

var port string

func StartServer() {
	Init()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	v1 := r.Group("/api/v1")
	AddRoutes(v1)
	fmt.Printf("Gin server listening on port %s\n", port)
	err := r.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func Init() {
	fmt.Println("Initializing Gin server...")
	err := gotenv.Load("gin-server/.env")
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
}
