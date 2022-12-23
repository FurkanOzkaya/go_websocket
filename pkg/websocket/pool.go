package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/furkanozkaya/socket_go_v1.0/pkg/common"
	"github.com/gorilla/websocket"
)

type ConnectionModel struct {
	Operation   string          `json:"operation,omitempty"`
	MessageType string          `json:"message_type,omitempty"`
	MessageId   string          `json:"message_id,omitempty"`
	User        string          `json:"user,omitempty"`
	From        string          `json:"from,omitempty"`
	To          string          `json:"to,omitempty"`
	Message     string          `json:"message,omitempty"`
	Status      string          `json:"status,omitempty"`
	Conn        *websocket.Conn `json:"conn,omitempty"`
}

type Pool struct {
	Operation chan *ConnectionModel
	Clients   map[string]*websocket.Conn
}

func NewPool() *Pool {
	return &Pool{
		Operation: make(chan *ConnectionModel),
		Clients:   make(map[string]*websocket.Conn),
	}
}

func (pool *Pool) Start() {

	var redis *common.Redis = common.GetRedis()

	for {
		client := <-pool.Operation
		switch client.Operation {
		case CONNECT:
			if pool.Clients == nil {
				pool.Clients = make(map[string]*websocket.Conn)
			}
			pool.Clients[client.User] = client.Conn
			send_msg := ConnectionModel{Operation: CONNECT, Message: "success"}
			go sendMessage(client.Conn, send_msg, pool)
			var ipAddress string = common.GetOutboundIP()
			log.Println(ipAddress)
			// TODO for now time is 6000 change to 60 for heartbeat
			writeRedisClient(redis, client.User, ipAddress, 6000*time.Second)

		case MESSAGE:
			log.Println("[Start] pool.MESSAGE chan to message sendMessage")
			value, ok := pool.Clients[client.To]
			var targetUser *websocket.Conn
			if ok {
				log.Println("Target User Found in System Target user")
				targetUser = value

				//targetUser := pool.Clients[client.To]
				//sendMsg := ConnectionModel{Operation: MESSAGE, From: client.From, To: client.To, Message: client.Message}
				go sendMessage(targetUser, *client, pool)
			} else if client.MessageType != REDIRECT {
				log.Println("Target User Not Found in System")
				writeRedisMessage(redis, client, 6000000*time.Second)

			}
		case STATUS:
			log.Println("[Start] pool.STATUS chan to message sendMessage")
			value, ok := pool.Clients[client.To]
			var targetUser *websocket.Conn
			if ok {
				targetUser = value
				go sendMessage(targetUser, *client, pool)
			} else {
				log.Println("Target User Not Found in System")
				writeRedisMessage(redis, client, 6000000*time.Second)
			}
		case DISCONNECT:
			delete(pool.Clients, client.User)
			send_msg := ConnectionModel{Operation: DISCONNECT, Message: "success"}
			delRedisClient(redis, client.User)
			go sendMessage(client.Conn, send_msg, pool)

		}
	}
}

func sendMessage(conn *websocket.Conn, msg ConnectionModel, pool *Pool) {
	log.Println("message will send ", msg)
	if err := conn.WriteJSON(msg); err != nil {
		log.Println(err)
		return
	}
	// send sent acknowledge to user
	// this will send if there is no eror (client connected and message will recieved)
	if msg.Operation == MESSAGE {
		acknowledge_msg := ConnectionModel{Operation: STATUS, Status: "sent", From: msg.To, To: msg.From, MessageId: msg.MessageId}
		log.Println("Acknowledge message will send ", acknowledge_msg)
		pool.Operation <- &acknowledge_msg
	}
}

func writeRedisClient(redis *common.Redis, user string, ipAddress string, ttl time.Duration) {
	// write informations as json if needed
	err := redis.Client.Set(user, ipAddress, ttl).Err()
	if err != nil {
		log.Println("Redis Key is not generated [writeRedisClient]", err)
	}
}

func writeRedisMessage(redis *common.Redis, value *ConnectionModel, ttl time.Duration) {
	// write informations as json if needed
	key, err1 := redis.Client.Get(value.To).Result()

	if err1 == nil {
		value2, err2 := json.Marshal(value)
		if err2 != nil {
			log.Println("Redis Key is not generated [writeRedisMessage]", err2)
		}

		log.Println("Message will send to", key, value)
		err3 := redis.Client.Publish(string(key), value2).Err()
		if err3 != nil {
			log.Println("Redis Key is not generated [writeRedisMessage]", err3)
		}
	}

}

func delRedisClient(redis *common.Redis, user string) {
	err := redis.Client.Del(user).Err()
	if err != nil {
		log.Println("Redis Key is not deleted")
	}
}
