package ginserver

import "github.com/gin-gonic/gin"

func AddRoutes(r *gin.RouterGroup) {

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Gin server is running!"})
	})

	r.GET("/ws", handleWebSocket)

}
