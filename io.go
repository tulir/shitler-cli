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
	"strings"
)

func receivePreJoin(data map[string]interface{}) bool {
	success, ok := data["success"].(bool)
	if !ok {
		printOutput(g, "Invalid map", data)
	} else if success {
		*authtoken = data["authtoken"].(string)
		status.Clear()
		fmt.Fprintln(status, "In game", data["game"])
		printOutput(g, "Successfully joined", data["game"])
		return true
	} else {
		msg, ok := data["message"].(string)
		if !ok {
			printOutput(g, "Invalid map", data)
			return false
		}
		switch msg {
		case "gamenotfound":
			printOutput(g, "Could not find the given game!")
		case "gamestarted":
			printOutput(g, "That game has already started (try giving your authtoken?)")
		case "full":
			printOutput(g, "That game is full (try giving your authtoken?)")
		case "nameused":
			printOutput(g, "The name", *name, "is already in use (try giving your authtoken?)")
		case "invalidname":
			printOutput(g, "Your name contains invalid characters or is too short or long")
		default:
			printOutput(g, "Unknown error:", data["message"].(string))
		}
	}
	return false
}

func receive(typ string, data map[string]interface{}) {
	switch typ {
	case "chat":
		printOutputf(g, "<%s> %s\n", data["sender"], data["message"])
	case "join":
		printOutput(g, data["name"], "joined the game.")
	case "quit":
		printOutput(g, data["name"], "left the game.")
	case "connected":
		printOutput(g, data["name"], "dataonnected.")
	case "disconnected":
		printOutput(g, data["name"], "disconnected.")
	case "start":
		role, _ := data["role"].(string)
		if role == "hitler" {
			printOutput(g, "The game has started. You're Hitler!")
		} else {
			printOutput(g, "The game has started. You're a", role)
		}

		ps, ok := data["players"].(map[string]interface{})
		playerList = make(map[string]string)
		for name, role := range ps {
			r, _ := role.(string)
			playerList[name] = r
		}
		if ok {
			setPlayerList(g, normalizePlayers(playerList))
		} else {
			setPlayerList(g, "Failed to load players")
		}
	default:
		printOutput(g, "Unidentified message from server:", data)
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
		msg["index"] = args[0]
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
