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
