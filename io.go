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
	"github.com/jroimartin/gocui"
	"strconv"
	"strings"
)

func receivePreJoin(data map[string]interface{}) bool {
	success, ok := data["success"].(bool)
	if !ok {
		printOutput("Invalid map", data)
	} else if success {
		*authtoken = data["authtoken"].(string)
		status.Clear()
		fmt.Fprintln(status, "In game", data["game"])
		printOutput("Successfully joined", data["game"])
		return true
	} else {
		msg, ok := data["message"].(string)
		if !ok {
			printOutput("Invalid map", data)
			return false
		}
		switch msg {
		case "gamenotfound":
			printOutput("Could not find the given game!")
		case "gamestarted":
			printOutput("That game has already started (try giving your authtoken?)")
		case "full":
			printOutput("That game is full (try giving your authtoken?)")
		case "nameused":
			printOutput("The name", *name, "is already in use (try giving your authtoken?)")
		case "invalidname":
			printOutput("Your name contains invalid characters or is too short or long")
		default:
			printOutput("Unknown error:", data["message"].(string))
		}
	}
	return false
}

func receive(typ string, data map[string]interface{}) {
	handler, ok := recHandlers[typ]
	if !ok {
		printOutput("Unidentified message from server:", data)
	} else {
		handler(data)
	}
}

func onInput(g *gocui.Gui, v *gocui.View) (nilrror error) {
	nilrror = nil
	var msg = make(map[string]string)
	data := strings.TrimSpace(v.Buffer())

	if !strings.HasPrefix(data, "/") {
		msg["type"] = "chat"
		msg["message"] = data
		v.Clear()
		conn.ch <- msg
		return
	}

	args := strings.Split(data, " ")
	command := strings.ToLower(args[0])[1:]
	args = args[1:]
	v.Clear()

	msg["type"] = command
Switch:
	switch command {
	case "create":
		go createGame()
		return
	case "chancellor":
		msg["type"] = "pickchancellor"
		msg["name"] = args[0]
	case "vote":
		switch strings.ToLower(args[0]) {
		case "ja", "yes", "1", "true":
			msg["vote"] = "ja"
		case "nein", "no", "0", "false":
			msg["vote"] = "nein"
		}
	case "start":
		break
	case "discard":
		discard, err := strconv.Atoi(args[0])
		if err != nil {
			for i, c := range discarding {
				if c == strings.ToLower(args[0]) {
					msg["index"] = strconv.Itoa(i)
					break Switch
				}
			}
			printOutput("There are no", args[0], "cards to discard")
		} else if discard > len(discarding) || discard <= 0 {
			printOutput("Invalid discard index", discard)
		} else {
			msg["index"] = strconv.Itoa(discard - 1)
		}
	case "veto":
		switch strings.ToLower(args[0]) {
		case "request", "ask":
			msg["type"] = "vetorequest"
		case "accept", "yes":
			msg["type"] = "vetoaccept"
		}
	case "president":
		msg["type"] = "presidentselect"
		fallthrough
	case "investigate", "execute":
		msg["name"] = args[0]
	case "join":
		msg["game"] = args[0]
		msg["name"] = *name
		msg["authtoken"] = *authtoken
	default:
		fmt.Fprintf(output, "Unknown command: %s\n", command)
		return
	}

	conn.ch <- msg
	return
}
