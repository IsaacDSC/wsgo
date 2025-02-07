package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/websocket"
)

func main() {
	url := "ws://localhost:80/ws"

	conn, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer conn.Close()
	fmt.Println("Conectado ao servidor WebSocket")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		for {
			var msg string
			err := websocket.Message.Receive(conn, &msg)
			if err != nil {
				log.Println("Erro ao receber mensagem:", err)
				return
			}
			fmt.Println("Mensagem recebida:", msg)
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			msg := fmt.Sprintf("Mensagem enviada em: %s", t.Format(time.RFC3339))
			err := websocket.Message.Send(conn, msg)
			if err != nil {
				log.Println("Erro ao enviar mensagem:", err)
				return
			}
			fmt.Println("Mensagem enviada:", msg)

		case <-interrupt:
			fmt.Println("Encerrando conexão...")
			return
		}
	}
}
