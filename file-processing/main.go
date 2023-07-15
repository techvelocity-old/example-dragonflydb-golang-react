package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"time"
)

const (
	redisQueuePrefix = "fileStatusQueue:"
	redisStatus      = "completed"
)

func main() {
	// Initialize Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // Add password if required
		DB:       0,  // Select appropriate Redis database
	})

	// Initialize Gin router
	router := gin.Default()

	// API route to handle file uploads
	router.POST("/upload", func(c *gin.Context) {
		// Get the file from the form data
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Simulate file processing time
		time.Sleep(time.Second * 10)

		// Get the userID from the form data
		userID := c.PostForm("userID")

		// Create the user-specific Redis queue key
		queueKey := redisQueuePrefix + userID

		// Push the file upload task to the user-specific Redis queue
		err = client.RPush(queueKey, file.Filename).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue file"})
			return
		}

		// Publish the "processing" status to the user-specific Redis queue
		err = client.RPush(queueKey, redisStatus).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish status"})
			return
		}

	})
	// Run the Gin server
	err := router.Run(":8081")
	if err != nil {
		log.Fatal(err)
	}
}
