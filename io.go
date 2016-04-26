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
	"bytes"
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
	switch typ {
	case "chat":
		printOutputf("<%s> %s\n", data["sender"], data["message"])
	case "join":
		printOutput(data["name"], "joined the game.")
	case "quit":
		printOutput(data["name"], "left the game.")
	case "connected":
		printOutput(data["name"], "connected.")
	case "disconnected":
		printOutput(data["name"], "disconnected.")
	case "start":
		role, _ := data["role"].(string)
		if role == "hitler" {
			printOutput("The game has started. You're Hitler!")
		} else {
			printOutput("The game has started. You're a", role)
		}

		ps, ok := data["players"].(map[string]interface{})
		playerList = make(map[string]string)
		for n, r := range ps {
			rs, _ := r.(string)
			if n == *name {
				rs = role
			}
			playerList[n] = rs
		}
		if ok {
			setPlayerList(normalizePlayers(playerList))
		} else {
			setPlayerList("Failed to load players")
		}
	case "president":
		name, _ := data["name"].(string)
		printOutput(name, "is now the president")
		setStatus(name, " is choosing a chancellor")
	case "startvote":
		president, _ := data["president"].(string)
		chancellor, _ := data["chancellor"].(string)
		printOutput(president, "has chosen", chancellor, "as the chancellor.")
		setStatus("Voting for president ", president, " and chancellor ", chancellor)
	case "vote":
		vote, _ := data["vote"].(string)
		printOutputf("You voted %s!\n", strings.Title(vote))
	case "cards":
		cs, _ := data["cards"].([]interface{})
		discarding = make([]string, len(cs))
		var buf bytes.Buffer
		buf.WriteString("Available cards - ")
		for i, c := range cs {
			card, _ := c.(string)
			discarding[i] = card
			buf.WriteString(fmt.Sprintf("%d: %s", i+1, card))
			if i != len(cs)-1 {
				buf.WriteString(", ")
			}
		}
		printOutput(buf.String())
	case "presidentdiscard":
		name, _ := data["name"].(string)
		printOutput("The president is discarding a policy...")
		setStatus(name, " to discard a policy")
	case "chancellordiscard":
		name, _ := data["name"].(string)
		printOutput("The chancellor is discarding a policy...")
		setStatus(name, " to discard a policy")
	case "table":
		deck, _ := data["deck"].(float64)
		discarded, _ := data["discarded"].(float64)
		tableLiberal, _ := data["tableLiberal"].(float64)
		tableFascist, _ := data["tableFascist"].(float64)
		setTable(fmt.Sprintf("Deck: %d\nDiscarded: %d\n\nEnacted\n  Liberal: %d\n  Fascist: %d\n", int(deck), int(discarded), int(tableLiberal), int(tableFascist)))
	case "enact":
		president, _ := data["president"].(string)
		chancellor, _ := data["chancellor"].(string)
		policy, _ := data["policy"].(string)
		printOutput(president, "and", chancellor, "have enacted a", policy, "policy.")
	case "forceenact":
		policy, _ := data["policy"].(string)
		printOutput("Three governments have failed and the frustrated populace has taken matters into their own hands, enacting a", policy, "policy.")
	case "peek":
		printOutput("The president has peeked at the next three cards.")
	case "peekcards":
		cs, _ := data["cards"].([]interface{})
		var buf bytes.Buffer
		buf.WriteString("The next three cards are ")
		for i, c := range cs {
			card, _ := c.(string)
			buf.WriteString(card)
			if i != len(cs)-1 {
				buf.WriteString(", ")
			}
		}
		printOutput(buf.String())
	default:
		printOutput("Unidentified message from server:", data)
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
