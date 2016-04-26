// shitler-cli - A command-line client for shitlerd
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jroimartin/gocui"
	"net/url"
	"time"
)

var interrupt = make(chan bool, 1)

type connection struct {
	ws     *websocket.Conn
	ch     chan interface{}
	joined bool
}

var conn *connection

func connect(g *gocui.Gui) error {
	fmt.Fprintf(status, "Connecting to %s\n", *address)

	u := url.URL{Scheme: protocolWS, Host: *address, Path: "/socket"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Fprintf(status, "Failed to connect: %v\n", err)
	}

	conn = &connection{ws: c, ch: make(chan interface{}), joined: false}
	go conn.writeLoop()
	go conn.readLoop()
	return nil
}

func (c *connection) Close() {
	c.write(websocket.CloseMessage, []byte{})
}

func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.ws.WriteMessage(mt, payload)
}

func (c *connection) writeJSON(payload interface{}) error {
	c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.ws.WriteJSON(payload)
}

func (c *connection) readLoop() {
	defer c.Close()
	for {
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			fmt.Fprintf(status, "Disconnected: %v\n", err)
			return
		}
		var rec = make(map[string]interface{})
		err = json.Unmarshal(data, &rec)
		if err != nil {
			printOutput(err)
		}
		if !c.joined {
			c.joined = receivePreJoin(rec)
			continue
		}
		typ, ok := rec["type"].(string)
		if !ok {
			printOutput("Invalid message from server:", rec)
		}
		receive(typ, rec)
	}
}

func (c *connection) writeLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case new, ok := <-conn.ch:
			if !ok {
				c.Close()
				return
			}
			err := c.writeJSON(new)
			if err != nil {
				setStatus("Disconnected:", err)
				return
			}
		case <-ticker.C:
			err := c.write(websocket.PingMessage, []byte{})
			if err != nil {
				setStatus("Disconnected:", err)
				return
			}
		case <-interrupt:
			c.Close()
			return
		}
	}
}
