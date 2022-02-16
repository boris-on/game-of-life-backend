package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boris-on/game-of-life-backend/game"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	pongWait = 120 * time.Second

	pingPeriod = (pongWait * 9) / 10
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	hub *Hub

	conn *websocket.Conn

	send chan []byte
}

func (c *Client) readPump(world *game.World) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	// c.conn.SetReadLimit(maxMessageSize)
	ctx := context.Background()

	for {
		_, msg, err := c.conn.Read(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println(msg)
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		c.hub.broadcast <- msg

		var event game.Event
		json.Unmarshal(msg, &event)
		world.HandleEvent(&event)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	ctx := context.Background()

	for {
		select {
		case msg, ok := <-c.send:

			if !ok {
				return
			}

			w, err := c.conn.Writer(ctx, websocket.MessageText)
			if err != nil {
				fmt.Println(err)
				return
			}

			w.Write(msg)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				fmt.Println(err)
				return
			}
		case <-ticker.C:
			if err := c.conn.Write(ctx, websocket.MessageBinary, nil); err != nil {
				fmt.Println(err)
				return
			}
		}
	}

}

func ServeWs(hub *Hub, world *game.World, w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	conn.SetReadLimit(524288000)

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	client.hub.register <- client

	unit := world.AddUnit()

	ctx := context.Background()
	err = wsjson.Write(ctx, conn, game.Event{
		Type: game.EventTypeInit,
		Data: game.EventInit{
			ID:    unit.ID,
			Units: world.Units,
			Area:  world.Area,
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	msg, err := json.Marshal(game.Event{
		Type: game.EventTypeConnect,
		Data: game.EventConnect{
			Unit: *unit,
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	hub.broadcast <- msg

	go client.writePump()
	go client.readPump(world)
}
