package ginserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func Emit(payload string) bool {
	connectionsMu.Lock()
	defer connectionsMu.Unlock()

	allSuccess := true
	for ip, conn := range connections {
		err := conn.WriteMessage(websocket.TextMessage, []byte(payload))
		if err != nil {
			fmt.Println("Write error to", ip, ":", err)
			conn.Close()
			delete(connections, ip)
			allSuccess = false
		}
	}
	return allSuccess
}

func AddRoutes(r *gin.RouterGroup) {

	// check server online
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Gin server is running!"})
	})

	// upload and split file
	r.POST("/seed", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
			return
		}

		seedFilePath := filepath.Join(baseTorrentPath, "files", filepath.Base(file.Filename))

		err = c.SaveUploadedFile(file, seedFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		metadata, err := splitFile(seedFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to split file"})
			return
		}

		metaFilePath := filepath.Join(baseTorrentPath, "metadata", file.Filename+".meta.json")
		metaFile, err := os.Create(metaFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create metadata file"})
			return
		}
		defer metaFile.Close()

		err = json.NewEncoder(metaFile).Encode(metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write metadata file"})
			return
		}

		Emit(fmt.Sprintf(`{"event":"uploaded","data":"%s"}`, file.Filename))

		c.JSON(http.StatusOK, gin.H{
			"message":  "File uploaded and split successfully",
			"metadata": metadata,
		})
	})

	// get metadata file
	r.GET("/metadata/:filename", func(c *gin.Context) {
		file := c.Param("filename")
		metaFilePath := filepath.Join(baseTorrentPath, "metadata", file+".meta.json")
		c.File(metaFilePath)
	})

	// get all metadata files
	r.GET("/metadata", func(c *gin.Context) {
		metaDir := "data/metadata"
		files, err := os.ReadDir(metaDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read metadata directory"})
			return
		}

		var allMetadata []ChunkMetadata

		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".json" && filepath.Ext(file.Name()) != ".meta.json" {
				continue
			}

			path := filepath.Join(metaDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue // skip unreadable files
			}

			var metadata ChunkMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				continue // skip invalid JSON
			}

			allMetadata = append(allMetadata, metadata)
		}

		c.JSON(http.StatusOK, allMetadata)
	})
}

func WebSocketRoutes(r *gin.Engine) {

	// connect to WebSocket server
	r.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c)
	})

	// broadcast message to all connected clients
	r.POST("/broadcast", func(c *gin.Context) {
		var emit WSMessage
		if err := c.BindJSON(&emit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		payload := fmt.Sprintf(`{"event":"%s","data":"%s"}`, emit.Event, emit.Data)

		success := Emit(payload)

		c.JSON(http.StatusOK, gin.H{"broadcast": success})
	})

}

func AddLimitedConcurrencyRoutes(r *gin.RouterGroup) {

	// download file chunks
	r.GET("/chunk/:filename/:index", func(c *gin.Context) {
		file := c.Param("filename")
		index := c.Param("index")

		chunkPath := filepath.Join(baseTorrentPath, fmt.Sprintf("%s.chunk.%s", file, index))
		if _, err := os.Stat(chunkPath); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chunk not found"})
			return
		}

		c.File(chunkPath)
	})
}
