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
	"strings"
)

func recChat(data map[string]interface{}) {
	printOutputf("<%s> %s\n", data["sender"], data["message"])
}

func recJoin(data map[string]interface{}) {
	printOutput(data["name"], "joined the game.")
}

func recQuit(data map[string]interface{}) {
	printOutput(data["name"], "left the game.")
}

func recConnect(data map[string]interface{}) {
	printOutput(data["name"], "connected.")
}

func recDisconnect(data map[string]interface{}) {
	printOutput(data["name"], "disconnected.")
}

func recStart(data map[string]interface{}) {
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
}

func recPresident(data map[string]interface{}) {
	name, _ := data["name"].(string)
	printOutput(name, "is now the president")
	setStatus(name, " is choosing a chancellor")
}

func recStartVote(data map[string]interface{}) {
	president, _ := data["president"].(string)
	chancellor, _ := data["chancellor"].(string)
	printOutput(president, "has chosen", chancellor, "as the chancellor.")
	setStatus("Voting for president ", president, " and chancellor ", chancellor)
}

func recVote(data map[string]interface{}) {
	vote, _ := data["vote"].(string)
	printOutputf("You voted %s!\n", strings.Title(vote))
}

func recCards(data map[string]interface{}) {
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
}

func recPresidentDiscard(data map[string]interface{}) {
	name, _ := data["name"].(string)
	printOutput("The president is discarding a policy...")
	setStatus(name, " to discard a policy")
}

func recChancellorDiscard(data map[string]interface{}) {
	name, _ := data["name"].(string)
	printOutput("The chancellor is discarding a policy...")
	setStatus(name, " to discard a policy")
}

func recTable(data map[string]interface{}) {
	deck, _ := data["deck"].(float64)
	discarded, _ := data["discarded"].(float64)
	tableLiberal, _ := data["tableLiberal"].(float64)
	tableFascist, _ := data["tableFascist"].(float64)
	setTable(fmt.Sprintf("Deck: %d\nDiscarded: %d\n\nEnacted\n  Liberal: %d\n  Fascist: %d\n", int(deck), int(discarded), int(tableLiberal), int(tableFascist)))
}

func recEnact(data map[string]interface{}) {
	president, _ := data["president"].(string)
	chancellor, _ := data["chancellor"].(string)
	policy, _ := data["policy"].(string)
	printOutput(president, "and", chancellor, "have enacted a", policy, "policy.")
}

func recForceEnact(data map[string]interface{}) {
	policy, _ := data["policy"].(string)
	printOutput("Three governments have failed and the frustrated populace has taken matters into their own hands, enacting a", policy, "policy.")
}

func recPeek(data map[string]interface{}) {
	printOutput("The president has peeked at the next three cards.")
}

func recPeekCards(data map[string]interface{}) {
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
}

func recInvestigateResult(data map[string]interface{}) {
	name, _ := data["name"].(string)
	role, _ := data["result"].(string)
	printOutput(name, "is a", role)
	playerList[name] = role
	setPlayerList(normalizePlayers(playerList))
}

func recInvestigate(data map[string]interface{}) {
	president, _ := data["president"].(string)
	printOutput("The president will now investigate a player.")
	setStatus(president, " is choosing a player to investigate")
}

func recPresidentSelect(data map[string]interface{}) {
	president, _ := data["president"].(string)
	printOutput("The president will now select the next president.")
	setStatus(president, " is selecting the next president")
}

func recExecute(data map[string]interface{}) {
	president, _ := data["president"].(string)
	printOutput("The president will now execute a player.")
	setStatus(president, " is choosing a player to execute")
}

func recInvestigated(data map[string]interface{}) {
	president, _ := data["president"].(string)
	name, _ := data["name"].(string)
	printOutput(president, "has investigated", name)
}

func recPresidentSelected(data map[string]interface{}) {
	president, _ := data["president"].(string)
	name, _ := data["name"].(string)
	printOutput(president, "has chosen", name, "as the special president.")
}

func recExecuted(data map[string]interface{}) {
	president, _ := data["president"].(string)
	name, _ := data["name"].(string)
	printOutput(president, "has executed", name+".")
}

func recError(data map[string]interface{}) {
	message, _ := data["message"].(string)
	printOutput("The server has encountered an internal error:", message)
}

func recEnd(data map[string]interface{}) {
	winner, _ := data["winner"].(string)
	printOutput("Game over!", strings.Title(winner)+"s", "won!")
	setStatus("Game over!")

	ps, ok := data["roles"].(map[string]interface{})
	playerList = make(map[string]string)
	for n, r := range ps {
		rs, _ := r.(string)
		playerList[n] = rs
	}
	if ok {
		setPlayerList(normalizePlayers(playerList))
	} else {
		setPlayerList("Failed to load players")
	}
}
