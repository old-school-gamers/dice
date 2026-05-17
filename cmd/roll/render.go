package main

import (
	"fmt"
	"io"

	"github.com/oldschoolgamers/dice"
)

// Render writes a human-readable representation of res to w.
// When color is true, ANSI escape codes are used to highlight kept dice,
// dim dropped/rerolled dice, and color the total green/red on hit/miss.
func Render(w io.Writer, color bool, res *dice.Result) {
	c := palette(color)

	_, _ = fmt.Fprintf(w, "%s = %s", res.Expression, c.total(res))
	if res.Target != nil {
		if res.Hit {
			_, _ = fmt.Fprintf(w, " %sHIT%s", c.green+c.bold, c.reset)
		} else {
			_, _ = fmt.Fprintf(w, " %sMISS%s", c.red+c.bold, c.reset)
		}
	}
	_, _ = fmt.Fprintln(w)

	for _, g := range res.Groups {
		_, _ = fmt.Fprintf(w, "  %s:", g.Notation)
		for _, d := range g.Dice {
			_, _ = fmt.Fprintf(w, " %s", c.die(d))
		}
		_, _ = fmt.Fprintln(w)
	}
}

type colors struct {
	reset, bold, dim, green, red, cyan string
}

func palette(enable bool) colors {
	if !enable {
		return colors{}
	}
	return colors{
		reset: "\x1b[0m",
		bold:  "\x1b[1m",
		dim:   "\x1b[2m",
		green: "\x1b[32m",
		red:   "\x1b[31m",
		cyan:  "\x1b[36m",
	}
}

func (c colors) die(d dice.Die) string {
	s := fmt.Sprintf("%d", d.Value)
	switch d.Status {
	case dice.StatusDropped:
		// Dimmed and bracketed so dropped dice are still readable in plain text.
		return c.dim + "[" + s + "]" + c.reset
	case dice.StatusRerolled:
		return c.dim + "(" + s + ")" + c.reset
	case dice.StatusExploded:
		return c.cyan + s + "!" + c.reset
	case dice.StatusSuccess:
		return c.green + c.bold + s + "*" + c.reset
	case dice.StatusFailure:
		return c.red + s + "x" + c.reset
	}
	return s
}

func (c colors) total(res *dice.Result) string {
	s := fmt.Sprintf("%d", res.Total)
	if res.Target != nil {
		if res.Hit {
			return c.green + c.bold + s + c.reset
		}
		return c.red + c.bold + s + c.reset
	}
	return c.bold + s + c.reset
}
