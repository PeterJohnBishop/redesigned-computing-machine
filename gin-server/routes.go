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

func AddRoutes(r *gin.RouterGroup) {

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Gin server is running!"})
	})

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

		c.JSON(http.StatusOK, gin.H{
			"message":  "File uploaded and split successfully",
			"metadata": metadata,
		})
	})

	r.GET("/metadata/:filename", func(c *gin.Context) {
		file := c.Param("filename")
		metaFilePath := filepath.Join(baseTorrentPath, "metadata", file+".meta.json")
		c.File(metaFilePath)
	})

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

	r.GET("/chunk/:filename/:index", func(c *gin.Context) {
		file := c.Param("filename")
		index := c.Param("index")

		chunkPath := filepath.Join(baseTorrentPath, "chunks", fmt.Sprintf("%s.chunk.%s", file, index))
		fmt.Println("Looking for chunk at:", chunkPath)
		if _, err := os.Stat(chunkPath); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chunk not found"})
			return
		}

		c.File(chunkPath)
	})

}

func WebSocketRoutes(r *gin.Engine) {

	r.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c)
	})

	r.POST("/broadcast", func(c *gin.Context) {
		var body struct {
			Message string `json:"message"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		for conn := range connections {
			err := conn.WriteMessage(websocket.TextMessage, []byte(body.Message))
			if err != nil {
				fmt.Println("Write error:", err)
				conn.Close()
				delete(connections, conn)
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "message broadcasted"})
	})

}
