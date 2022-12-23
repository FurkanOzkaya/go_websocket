package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/furkanozkaya/socket_go_v1.0/pkg/websocket"
)

// define our WebSocket endpoint
func serveWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)

	ws, err := websocket.Upgrade(w, r)
	if err != nil {
		log.Println(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	go websocket.Reader(pool, ws) // make it with goroutine # TODO
	go websocket.RedisReader(pool)
}

func setupRoutes() {
	pool := websocket.NewPool()
	go pool.Start()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})
}

func main() {
	fmt.Println("Real Time Server v1.0")
	setupRoutes()
	http.ListenAndServe(":8088", nil)
}
