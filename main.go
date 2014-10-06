package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: nozzle <port>")
		os.Exit(1)
	}

	port := fmt.Sprintf(":%s", os.Args[1])
	log.Printf("Starting WebSocket server on port %v\n", port)

	subscribers := make(chan chan []byte)
	go readStdin(subscribers)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, subscribers)
	})
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalln(err)
	}
}

func readStdin(subscribers chan chan []byte) {
	allSubscribers := make([]chan []byte, 0)

	stdin := make(chan []byte)
	go func(stdin chan []byte) {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				log.Fatalln(err)
			}
			stdin <- line
		}
	}(stdin)

	for {
		select {
		case subscription := <-subscribers:
			allSubscribers = append(allSubscribers, subscription)
		case line := <-stdin:
			activeSubscribers := make([]chan []byte, 0)
			for i := range allSubscribers {
				_, ok := <-allSubscribers[i]
				if !ok {
					continue
				} else {
					allSubscribers[i] <- line
					activeSubscribers = append(activeSubscribers, allSubscribers[i])
				}
			}
			allSubscribers = activeSubscribers
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request, subscribers chan chan []byte) {
	conn, err := websocket.Upgrade(w, r, w.Header(), 4096, 4096)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Incoming connection: %s\n", r.RemoteAddr)
	subscription := make(chan []byte)
	subscribers <- subscription
	subscription <- []byte("acc")
	defer close(subscription)
	defer log.Printf("Connection closed: %s\n", r.RemoteAddr)

	log.Printf("Stream started: STDIN -> %v\n", r.RemoteAddr)

	for {
		line := <-subscription
		if err = conn.WriteMessage(websocket.TextMessage, line); err != nil {
			log.Println(err)
			return
		}
		subscription <- []byte("acc")
	}
}
