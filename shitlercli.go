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
	"flag"
	"fmt"
	"github.com/jroimartin/gocui"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var address = flag.String("address", "localhost:29305", "The address of the shitler server.")
var secure = flag.Bool("secure", false, "Use secure connections (https/wss)")
var name = flag.String("name", "CLI-Guest", "The name to join with.")
var authtoken = flag.String("authtoken", "", "Auth token to retake username.")
var protocolHTTP = "http"
var protocolWS = "ws"

var status, output, players, input, errView *gocui.View

func main() {
	flag.Parse()

	if *secure {
		protocolHTTP += "s"
		protocolWS += "s"
	}

	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		panic(err)
	}
	defer g.Close()

	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		interrupt <- true
		return gocui.ErrQuit
	})

	g.SetLayout(layout)
	g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, onInput)

	g.Execute(load)
	g.Execute(connect)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

func onInput(g *gocui.Gui, v *gocui.View) (nilrror error) {
	nilrror = nil
	var msg = make(map[string]string)

	args := strings.Split(strings.TrimSpace(v.Buffer()), " ")
	command := strings.ToLower(args[0])
	args = args[1:]
	v.Clear()

	msg["type"] = command
	switch command {
	case "create":
		u := url.URL{Scheme: protocolHTTP, Host: *address, Path: "/create"}
		resp, err := http.DefaultClient.Get(u.String())
		if err != nil {
			fmt.Fprintln(output, "Failed to create game:", err)
			return
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(output, "Failed to create game:", err)
			return
		}
		fmt.Fprintln(output, "Created game:", string(data))
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

func load(g *gocui.Gui) error {
	status.Title = "Status"
	players.Title = "Players"
	output.Wrap = true
	input.Title = "Input"
	input.Editable = true
	return nil
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
