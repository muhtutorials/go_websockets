package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"time"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = pongWait * 9 / 10
)

func deadline() time.Time {
	return time.Now().Add(pongWait)
}

type Client struct {
	conn    *websocket.Conn
	manager *Manager
	room    string
	// event is used to avoid concurrent writes to WS connection
	event chan Event
}

func NewClient(c *websocket.Conn, m *Manager) *Client {
	return &Client{
		conn:    c,
		manager: m,
		event:   make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer c.manager.removeClient(c)

	c.conn.SetReadLimit(512)
	c.conn.SetPongHandler(c.pongHandler)

	if err := c.conn.SetReadDeadline(deadline()); err != nil {
		log.Println(err)
		return
	}

	for {
		var event Event

		err := c.conn.ReadJSON(&event)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("error reading message:", err)
			}
			log.Println(err)
			return
		}

		if err = c.manager.routeEvent(event, c); err != nil {
			log.Println(err)
			return
		}
	}
}

func (c *Client) writeMessages() {
	defer c.manager.removeClient(c)

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case event, ok := <-c.event:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					fmt.Println("connection closed:", err)
				}
				return
			}

			if err := c.conn.WriteJSON(event); err != nil {
				fmt.Println("failed to send message:", err)
			}
		case <-ticker.C:
			log.Println("ping")
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Println("failed to send ping:", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(msg string) error {
	log.Println("pong")
	return c.conn.SetReadDeadline(deadline())
}

type Clients map[*Client]struct{}

func (c Clients) String() string {
	var addrs []string
	for client := range c {
		addrs = append(addrs, client.conn.RemoteAddr().String())
	}
	return strings.Join(addrs, ", ")
}
