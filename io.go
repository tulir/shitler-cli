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

var recHandlers map[string]func(data map[string]interface{})

func init() {
	recHandlers = make(map[string]func(data map[string]interface{}))
	recHandlers["chat"] = recChat
	recHandlers["join"] = recJoin
	recHandlers["quit"] = recQuit
	recHandlers["connected"] = recConnect
	recHandlers["disconnected"] = recDisconnect
	recHandlers["start"] = recStart
	recHandlers["president"] = recPresident
	recHandlers["startvote"] = recStartVote
	recHandlers["vote"] = recVote
	recHandlers["cards"] = recCards
	recHandlers["presidentdiscard"] = recPresidentDiscard
	recHandlers["chancellordiscard"] = recChancellorDiscard
	recHandlers["table"] = recTable
	recHandlers["enact"] = recEnact
	recHandlers["forceenact"] = recForceEnact
	recHandlers["peek"] = recPeek
	recHandlers["peekcards"] = recPeekCards
	recHandlers["investigateresult"] = recInvestigateResult
	recHandlers["investigate"] = recInvestigate
	recHandlers["presidentselect"] = recPresidentSelect
	recHandlers["execute"] = recExecute
	recHandlers["investigated"] = recInvestigated
	recHandlers["presidentselected"] = recPresidentSelected
	recHandlers["executed"] = recExecuted
	recHandlers["end"] = recEnd
	recHandlers["error"] = recError
}

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

	switch command {
	case "create":
		go createGame()
		return
	case "chancellor":
		msg["type"] = "pickchancellor"
		msg["name"] = args[0]
	case "vote":
		msg["vote"] = cmdVote(args[0])
	case "start":
		break
	case "discard":
		msg["index"] = cmdDiscard(args[0])
	case "veto":
		msg["type"] = cmdVetoRequest(args[0])
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

func cmdVote(arg string) string {
	switch strings.ToLower(arg) {
	case "ja", "yes", "1", "true":
		return "ja"
	case "nein", "no", "0", "false":
		return "nein"
	default:
		return ""
	}
}

func cmdVetoRequest(arg string) string {
	switch strings.ToLower(arg) {
	case "request", "ask":
		return "vetorequest"
	case "accept", "yes":
		return "vetoaccept"
	default:
		return ""
	}
}

func cmdDiscard(arg string) string {
	discard, err := strconv.Atoi(arg)
	if err != nil {
		for i, c := range discarding {
			if c == strings.ToLower(arg) {
				return strconv.Itoa(i)
			}
		}
		printOutput("There are no", arg, "cards to discard")
		return ""
	} else if discard > len(discarding) || discard <= 0 {
		printOutput("Invalid discard index", discard)
		return ""
	}
	return strconv.Itoa(discard - 1)
}
