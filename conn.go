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
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jroimartin/gocui"
	"net/url"
	"time"
)

var interrupt = make(chan bool, 1)
var messages = make(chan map[string]string)

func connect(g *gocui.Gui) error {
	go func() {
		fmt.Fprintf(status, "Connecting to %s\n", *address)

		u := url.URL{Scheme: "ws", Host: *address, Path: "/socket"}

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			fmt.Fprintf(status, "Failed to connect: %v\n", err)
		}
		defer c.Close()

		done := make(chan struct{})

		go func() {
			defer c.Close()
			defer close(done)
			for {
				var data []byte
				_, data, err = c.ReadMessage()
				if err != nil {
					fmt.Fprintf(status, "Disconnected: %v\n", err)
					return
				}
				fmt.Fprintf(output, "Received %s\n", string(data))
			}
		}()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case new, ok := <-messages:
				if !ok {
					err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						return
					}
					c.Close()
					return
				}
				fmt.Fprintf(output, "Sending message %v...", new)
				err = c.WriteJSON(new)
				if err != nil {
					fmt.Fprintf(status, "Disconnected: %v\n", err)
					return
				}
				fmt.Fprint(output, " done!\n")
			case <-ticker.C:
				err = c.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					fmt.Fprintf(status, "Disconnected: %v\n", err)
					return
				}
			case <-interrupt:
				err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					return
				}
				c.Close()
				return
			}
		}
	}()
	return nil
}
