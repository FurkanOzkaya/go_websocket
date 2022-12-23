package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/furkanozkaya/socket_go_v1.0/pkg/common"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
		//delete user if its broken
	}

	return conn, nil
}

const (
	CONNECT    string = "connect"
	MESSAGE    string = "message"
	DISCONNECT string = "disconnect"
	REDIRECT   string = "redirect"
	STATUS     string = "status"
)

func Reader(pool *Pool, ws *websocket.Conn) {
	for {
		var wsmessage ConnectionModel
		err := ws.ReadJSON(&wsmessage)
		if err != nil {
			log.Println(err)
			return
		}
		if wsmessage.Operation == CONNECT || wsmessage.Operation == MESSAGE || wsmessage.Operation == DISCONNECT || wsmessage.Operation == STATUS {
			wsmessage.Conn = ws
			pool.Operation <- &wsmessage
		} else {
			log.Println("Wrong Operation Type")
		}

		fmt.Printf("Message Received: %+v\n", wsmessage)

	}
}

func RedisReader(pool *Pool) {
	var redis *common.Redis = common.GetRedis()
	var ipAddress string = common.GetOutboundIP()
	log.Println(ipAddress)
	sub := redis.Client.Subscribe(ipAddress)
	for {
		message, err := sub.ReceiveMessage()
		if err != nil {
			log.Println("Recieve Error")
		}
		wsmessage := ConnectionModel{}
		err1 := json.Unmarshal([]byte(message.Payload), &wsmessage)

		if err1 != nil {
			log.Println("Recieve Error")
		}

		fmt.Printf("Message Received: %+v\n", wsmessage)
		if wsmessage.Operation == CONNECT || wsmessage.Operation == MESSAGE || wsmessage.Operation == DISCONNECT || wsmessage.Operation == STATUS {
			log.Println("Message Received and sending to POOL again")
			wsmessage.MessageType = REDIRECT
			pool.Operation <- &wsmessage
		} else {
			log.Println("Wrong Operation Type")
		}

	}
}
