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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jroimartin/gocui"
	"unicode/utf8"
)

var address = flag.String("address", "localhost:29305", "The address of the shitler server.")

var status, output, players, input, errView *gocui.View

func main() {
	flag.Parse()

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

func layout(g *gocui.Gui) (err error) {
	maxX, maxY := g.Size()
	if maxX < 70 || maxY < 15 {
		errView, err = g.SetView("error", 0, 0, maxX-1, maxY-1)
		if err != nil && err != gocui.ErrUnknownView {
			return
		}
		errView.Wrap = true
		errView.Clear()
		fmt.Fprintf(errView, "Window too small!")
		return nil
	}

	if errView != nil {
		g.DeleteView(errView.Name())
		errView = nil
	}

	status, err = g.SetView("status", 0, 0, maxX-1, 2)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	output, err = g.SetView("output", 0, 3, maxX-29, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		return
	}

	players, err = g.SetView("players", maxX-28, 3, maxX-1, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		return
	}

	input, err = g.SetView("input", 0, maxY-3, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return
	}
	g.SetCurrentView("input")

	return nil
}

func onInput(g *gocui.Gui, v *gocui.View) error {
	var msg = make(map[string]string)
	dec := json.NewDecoder(v)
	dec.Decode(&msg)
	messages <- msg
	v.Clear()
	return nil
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
		if utf8.RuneCountInString(player) > maxLen {
			maxLen = utf8.RuneCountInString(player)
		}
	}
	var buf bytes.Buffer
	for player, role := range players {
		buf.WriteString(player)
		if utf8.RuneCountInString(player) < maxLen {
			for i := utf8.RuneCountInString(player); i < maxLen; i++ {
				buf.WriteString(" ")
			}
		}
		buf.WriteString(" - ")
		buf.WriteString(role)
		buf.WriteRune('\n')
	}
	return buf.String()
}
