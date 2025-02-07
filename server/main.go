package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

var redisClient = redis.NewClient(&redis.Options{
	Addr: "redis:6379",
})

var clients = make(map[*websocket.Conn]bool)
var mu sync.Mutex

const redisChannel = "chat_channel"

func websocketHandler(ws *websocket.Conn) {
	defer ws.Close()
	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	fmt.Println("Novo cliente conectado!")

	for {
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			fmt.Println("Cliente desconectado:", err)
			break
		}

		fmt.Println("Mensagem recebida:", msg)

		redisClient.Publish(context.Background(), redisChannel, msg)
	}

	mu.Lock()
	delete(clients, ws)
	mu.Unlock()
	fmt.Println("Cliente removido.")
}

func redisSubscriber() {
	pubsub := redisClient.Subscribe(context.Background(), redisChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		fmt.Println("Mensagem recebida do Redis:", msg.Payload)

		mu.Lock()
		for client := range clients {
			err := websocket.Message.Send(client, msg.Payload)
			if err != nil {
				fmt.Println("Erro ao enviar mensagem:", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

func main() {
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Erro ao conectar ao Redis:", err)
	}

	go redisSubscriber()

	http.Handle("/ws", websocket.Handler(websocketHandler))

	http.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	fmt.Println(os.Args)

	port := flag.Int("port", 8080, "Port to run the server on")

	flag.Parse()

	fmt.Println("Servidor WebSocket rodando na porta", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal("Erro ao iniciar o servidor:", err)
	}
}
