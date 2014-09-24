package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, w.Header(), 4096, 4096)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Incoming connection: %s\n", r.RemoteAddr)
	defer log.Printf("Connection closed: %s\n", r.RemoteAddr)

	_, streamEndpoint, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := http.Get(string(streamEndpoint))
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Stream started: %v -> %v\n", string(streamEndpoint), r.RemoteAddr)

	rd := bufio.NewReader(resp.Body)

	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			log.Println(err)
			return
		}

		if err = conn.WriteMessage(websocket.TextMessage, line); err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: nozzle <port>")
		os.Exit(1)
	}

	port := fmt.Sprintf(":%s", os.Args[1])
	log.Printf("Starting WebSocket server on port %v\n", port)

	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalln(err)
	}
}
