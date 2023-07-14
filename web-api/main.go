package main

import (
	//"doc-editor/db"
	"doc-editor/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var (
	redis_host = "localhost"
	redis_port = "27017"
	redis_uri  = fmt.Sprintf("redis://%s:%s/0", redis_host, redis_port)
)

func InitRedis() (*redis.Client, error) {
	opt, err := redis.ParseURL(redis_uri)

	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opt)
	return rdb, nil
}

func main() {
	rdb, err := InitRedis()
	if err != nil {
		log.Fatal("Couldn't connect to Redis:", err)
	}
	//pgdb := db.InitDB()
	router := gin.Default()
	document := NewDocument()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		},
	}

	router.GET("/api/ws", handleWebSocket(rdb, document, &upgrader))
	//router.GET("/api/docs", getDocs(pgdb))

	err = router.Run(":8000")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}

//func getDocs(pgdb *gorm.DB) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		var docs []models.Document
//		result := pgdb.Find(&docs)
//		c.JSON(200, result)
//	}
//}

func NewDocument() *models.Document {
	document := &models.Document{
		Clients: make(map[*websocket.Conn]bool),
	}
	return document
}

func handleWebSocket(
	rdb *redis.Client,
	document *models.Document,
	upgrader *websocket.Upgrader) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Error upgrading connection to WebSocket:", err)
			return
		}

		// Register client
		document.Lock()
		document.Clients[conn] = true
		document.Unlock()

		// Handle incoming messages
		go handleIncomingMessages(c, rdb, document, conn)
	}
}

func handleIncomingMessages(
	c *gin.Context,
	rdb *redis.Client,
	document *models.Document,
	conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)

			// Unregister client
			document.Lock()
			delete(document.Clients, conn)
			document.Unlock()

			// Close the connection
			conn.Close()
			break
		}

		document.Lock()
		if len(message) > 0 {
			document.Content = string(message)

			// Push the updated document to the Redis channel
			err := rdb.RPush(c.Request.Context(), "document_channel", document.Content).Err()
			if err != nil {
				log.Println("Error pushing document to Redis channel:", err)
			}
		} else {
			// Replace empty message with sample text
			document.Content = "Start typing ..."
		}
		document.Unlock()

		document.RLock()
		for client := range document.Clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(document.Content))
			if err != nil {
				log.Println("Error sending message to client:", err)
				client.Close()
				delete(document.Clients, client)
			}
		}
		document.RUnlock()
	}
}
