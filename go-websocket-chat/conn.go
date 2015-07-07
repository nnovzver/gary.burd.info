package main

import (
	"golang.org/x/net/websocket"
	"net/http"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *connection) reader() {
	for {
		message := make([]byte, 1024)
		n, err := c.ws.Read(message)
		if err != nil {
			break
		}
		Hub.broadcast <- message[:n]
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		_, err := c.ws.Write(message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

func wsHandler(ws *websocket.Conn) {
	c := &connection{send: make(chan []byte, 256), ws: ws}
	Hub.register <- c
	defer func() { Hub.unregister <- c }()
	go c.writer()
	c.reader()
}

func connHandler(w http.ResponseWriter, r *http.Request) {
	s := websocket.Server{Handler: wsHandler}
	s.ServeHTTP(w, r)
}
