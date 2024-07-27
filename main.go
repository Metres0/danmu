package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// 定义升级器，用于将 HTTP 连接升级为 WebSocket 连接
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 记录连接的客户端
var clients = make(map[*websocket.Conn]bool)

// 用于广播消息的通道
var broadcast = make(chan Message)

// 定义消息的结构体
type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

func main() {
	// 设置静态文件路由
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// 设置 WebSocket 路由
	http.HandleFunc("/ws", handleConnections)

	// 启动处理消息的协程
	go handleMessages()

	// 启动 HTTP 服务器
	log.Println("HTTP server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// 处理 WebSocket 连接
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// 升级 HTTP 连接为 WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// 注册新的客户端
	clients[ws] = true

	for {
		var msg Message
		// 读取消息
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// 将消息发送到广播通道
		broadcast <- msg
	}
}

// 处理来自广播通道的消息
func handleMessages() {
	for {
		// 从广播通道中接收消息
		msg := <-broadcast
		// 发送消息给所有已连接的客户端
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
