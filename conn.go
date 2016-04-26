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
	ws      *websocket.Conn
	g       *gocui.Gui
	ch      chan interface{}
	joined  bool
	players map[string]string
}

var conn *connection

func connect(g *gocui.Gui) error {
	fmt.Fprintf(status, "Connecting to %s\n", *address)

	u := url.URL{Scheme: protocolWS, Host: *address, Path: "/socket"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Fprintf(status, "Failed to connect: %v\n", err)
	}

	conn = &connection{ws: c, g: g, ch: make(chan interface{}), joined: false}
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
			fmt.Fprintln(output, err)
		}
		if !c.joined {
			success, ok := rec["success"].(bool)
			if !ok {
				fmt.Fprintln(output, "Invalid map", rec)
			} else if success {
				*authtoken = rec["authtoken"].(string)
				c.joined = true
				status.Clear()
				fmt.Fprintln(status, "In game", rec["game"])
				fmt.Fprintln(output, "Successfully joined", rec["game"])
			} else {
				msg, ok := rec["message"].(string)
				if !ok {
					fmt.Fprintln(output, "Invalid map", rec)
					continue
				}
				switch msg {
				case "gamenotfound":
					fmt.Fprintln(output, "Could not find the given game!")
				case "gamestarted":
					fmt.Fprintln(output, "That game has already started (try giving your authtoken?)")
				case "full":
					fmt.Fprintln(output, "That game is full (try giving your authtoken?)")
				case "nameused":
					fmt.Fprintln(output, "The name", *name, "is already in use (try giving your authtoken?)")
				case "invalidname":
					fmt.Fprintln(output, "Your name contains invalid characters or is too short or long")
				default:
					fmt.Fprintln(output, "Unknown error:", rec["message"].(string))
				}
			}
			continue
		}
		typ, ok := rec["type"].(string)
		if !ok {
			fmt.Fprintln(output, "Invalid message from server:", rec)
		}
		switch typ {
		case "chat":
			fmt.Fprintf(output, "<%s> %s", rec["sender"], rec["message"])
		case "join":
			fmt.Fprintln(output, rec["name"], "joined the game.")
		case "quit":
			fmt.Fprintln(output, rec["name"], "left the game.")
		case "connected":
			fmt.Fprintln(output, rec["name"], "reconnected.")
		case "disconnected":
			fmt.Fprintln(output, rec["name"], "disconnected.")
		case "start":
			fmt.Fprintln(output, "The game has started. Your role is", rec["role"])
			players.Clear()
			var ok bool
			c.players, ok = rec["players"].(map[string]string)
			if ok {
				fmt.Fprintln(players, normalizePlayers(c.players))
			} else {
				fmt.Fprintln(players, "Failed to load players")
			}
		default:
			fmt.Fprintln(output, "Unidentified message from server:", rec)
		}
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
				fmt.Fprintf(status, "Disconnected: %v\n", err)
				return
			}
		case <-ticker.C:
			err := c.write(websocket.PingMessage, []byte{})
			if err != nil {
				fmt.Fprintf(status, "Disconnected: %v\n", err)
				return
			}
		case <-interrupt:
			c.Close()
			return
		}
	}
}
