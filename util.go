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
	"io/ioutil"
	"net/http"
	"net/url"
)

func createGame() {
	u := url.URL{Scheme: protocolHTTP, Host: *address, Path: "/create"}
	resp, err := http.DefaultClient.Get(u.String())
	if err != nil {
		printOutput("Failed to create game:", err)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		printOutput("Failed to create game:", err)
		return
	}
	printOutput("Created game", string(data))
}

func normalizePlayers(players map[string]string) string {
	var maxLen int
	for player := range players {
		if len(player) > maxLen {
			maxLen = len(player)
		}
	}
	var buf bytes.Buffer
	for player, role := range players {
		buf.WriteString(player)
		if len(player) < maxLen {
			for i := len(player); i < maxLen; i++ {
				buf.WriteString(" ")
			}
		}
		buf.WriteString(" - ")
		buf.WriteString(role)
		buf.WriteRune('\n')
	}
	return buf.String()
}
