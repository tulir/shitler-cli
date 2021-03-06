// shitler-cli - A command-line client for shitlerd
// Copyright (C) 2016-2017 Tulir Asokan

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
	"flag"

	"github.com/jroimartin/gocui"
)

var address = flag.String("address", "localhost:29305", "The address of the shitler server.")
var secure = flag.Bool("secure", false, "Use secure connections (https/wss)")
var name = flag.String("name", "CLI-Guest", "The name to join with.")
var authtoken = flag.String("authtoken", "", "Auth token to retake username.")
var g *gocui.Gui

var lobbyPlayers map[string]bool
var playerList map[string]string
var discarding []string

var protocolHTTP = "http"
var protocolWS = "ws"

func main() {
	flag.Parse()

	if *secure {
		protocolHTTP += "s"
		protocolWS += "s"
	}

	var err error
	g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		interrupt <- true
		return gocui.ErrQuit
	})

	g.SetManagerFunc(layout)
	g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, onInput)

	g.Execute(load)
	g.Execute(connect)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}
