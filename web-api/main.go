package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

const (
	redisQueuePrefix = "fileStatusQueue:"
	redisStatus      = "processing"
	uploadAPIURL     = "http://localhost:8081/upload"
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

	// Apply CORS middleware
	router.Use(cors.Default())

	// API route to handle file uploads
	router.POST("/upload", func(c *gin.Context) {
		// Get the file from the form data
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get the userID from the form data
		userID := c.PostForm("userID")

		// Create the user-specific Redis queue key
		queueKey := redisQueuePrefix + userID

		// Push the file upload task to the user-specific Redis queue
		err = client.RPush(queueKey, file.Filename).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue file"})
			log.Println("Failed to enqueue file:", err)
			return
		}

		// Publish the "processing" status to the user-specific Redis queue
		err = client.RPush(queueKey, redisStatus).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish status"})
			log.Println("Failed to publish status:", err)
			return
		}

		// Send the file to the separate API
		go sendFileToAPI(file, userID)

		c.JSON(http.StatusOK, gin.H{"message": "File uploaded and processing"})
	})

	// Run the Gin server
	err := router.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

// Helper function to send the file to the separate API
func sendFileToAPI(file *multipart.FileHeader, userID string) {
	// Open the file
	srcFile, err := file.Open()
	if err != nil {
		log.Println("Failed to open file:", err)
		return
	}
	defer srcFile.Close()

	// Create a new multipart buffer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a new form-data field for the file
	fileWriter, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		log.Println("Failed to create form file:", err)
		return
	}

	// Copy the file data to the form-data field
	_, err = io.Copy(fileWriter, srcFile)
	if err != nil {
		log.Println("Failed to copy file data:", err)
		return
	}

	// Add the userID to the form-data
	_ = writer.WriteField("userID", userID)

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		log.Println("Failed to close multipart writer:", err)
		return
	}

	// Create a new HTTP POST request to the upload API
	req, err := http.NewRequest("POST", uploadAPIURL, body)
	if err != nil {
		log.Println("Failed to create HTTP request:", err)
		return
	}

	// Set the content-type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to send HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		log.Println("Unexpected response status:", resp.StatusCode)
		return
	}

	log.Println("File sent to API:", file.Filename)
}
