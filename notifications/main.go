package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

const (
	redisQueuePrefix = "fileStatusQueue:"
	pingInterval     = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var connections sync.Map

func main() {
	// Initialize Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // Add password if required
		DB:       0,  // Select appropriate Redis database
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Println("Error while sending Redis ping:", err)
	} else {
		log.Println("Redis ping response:", pong)
	}

	// Initialize Gin router
	router := gin.Default()

	// WebSocket route to handle file status updates
	router.GET("/notifications/ws/:userID", func(c *gin.Context) {
		userID := c.Param("userID")
		log.Println(userID)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}
		// Store the WebSocket connection
		connections.Store(userID, conn)

		// Close the WebSocket connection and remove it from the map when the client disconnects
		defer func() {
			conn.Close()
			connections.Delete(userID)

			// Remove the Redis queue channel
			queueKey := redisQueuePrefix + userID
			client.Del(queueKey)
		}()

		// Create the user-specific Redis queue
		queueKey := redisQueuePrefix + userID

		// Read messages from Redis queue and send file status updates to the WebSocket client
		for {
			result, err := client.BLPop(0, queueKey).Result()
			log.Println(result)
			if err != nil {
				log.Println("Error while popping item from Redis queue:", err)
				return
			}

			fileStatus := result[1]
			log.Println(fileStatus)

			// Send the file status update to the WebSocket client
			err = conn.WriteMessage(websocket.TextMessage, []byte(fileStatus))
			if err != nil {
				log.Println("WebSocket send error:", err)
				return
			}

			// Check if the sent file status update indicates completion
			// If yes, break the loop and close the WebSocket connection
			if isCompletionStatus(fileStatus) {
				break
			}
		}
	})

	// Run the Gin server
	err = router.Run(":8082")
	if err != nil {
		log.Fatal(err)
	}
}

// Helper function to check if the file status update indicates completion
func isCompletionStatus(fileStatus string) bool {
	// Customize this logic based on your file status update format
	// Return true if the status indicates completion, false otherwise
	return fileStatus == "completed"
}
