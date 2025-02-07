package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

// Configuração do Redis
var redisClient = redis.NewClient(&redis.Options{
	Addr: "redis:6379", // Altere para a URL do seu Redis
})

// Mapa de conexões WebSocket
var clients = make(map[*websocket.Conn]bool)
var mu sync.Mutex // Para sincronização do mapa

// Canal Redis
const redisChannel = "chat_channel"

// Handler WebSocket
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

		// Publica a mensagem no Redis para outras instâncias
		redisClient.Publish(context.Background(), redisChannel, msg)
	}

	// Remove o cliente ao desconectar
	mu.Lock()
	delete(clients, ws)
	mu.Unlock()
	fmt.Println("Cliente removido.")
}

// Escuta mensagens do Redis e retransmite para os WebSockets conectados
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

	// Inicia a escuta do Redis
	go redisSubscriber()

	// Configura o servidor WebSocket
	http.Handle("/ws", websocket.Handler(websocketHandler))

	// Inicia o servidor
	port := flag.Int("port", 8080, "Port to run the server on")
	//port := ":8080"

	// Parse flags
	flag.Parse()

	fmt.Println("Servidor WebSocket rodando na porta", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal("Erro ao iniciar o servidor:", err)
	}
}
