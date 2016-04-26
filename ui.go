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
)

func printOutput(msg ...interface{}) {
	g.Execute(func(g *gocui.Gui) error {
		fmt.Fprintln(output, msg...)
		return nil
	})
}

func printOutputf(msg string, args ...interface{}) {
	g.Execute(func(g *gocui.Gui) error {
		fmt.Fprintf(output, msg, args...)
		return nil
	})
}

func setStatus(msg ...interface{}) {
	g.Execute(func(g *gocui.Gui) error {
		status.Clear()
		fmt.Fprint(status, msg...)
		return nil
	})
}

func setPlayerList(list string) {
	g.Execute(func(g *gocui.Gui) error {
		players.Clear()
		fmt.Fprintln(players, list)
		return nil
	})
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
		return
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

func load(g *gocui.Gui) error {
	status.Title = "Status"
	players.Title = "Players"
	output.Title = "Output"
	output.Wrap = true
	output.Autoscroll = true
	input.Title = "Input"
	input.Editable = true
	return nil
}
