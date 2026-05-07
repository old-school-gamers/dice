package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/oldschoolgamers/dice"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: roll <expression>")
		fmt.Fprintln(os.Stderr, "examples:")
		fmt.Fprintln(os.Stderr, "  roll 2d6+3")
		fmt.Fprintln(os.Stderr, "  roll 4d6kh3")
		fmt.Fprintln(os.Stderr, "  roll 'adv(1d20)+5'")
		fmt.Fprintln(os.Stderr, "  roll 1d20+5 '>=' 15")
		fmt.Fprintln(os.Stderr, "  roll 5d10cs'>='8")
	}
	explain := flag.Bool("explain", false, "print a plain-English description and exit without rolling")
	flag.Parse()

	expr := strings.Join(flag.Args(), " ")
	if strings.TrimSpace(expr) == "" {
		flag.Usage()
		os.Exit(2)
	}

	if *explain {
		e, err := dice.Parse(expr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "roll: %v\n", err)
			os.Exit(2)
		}
		fmt.Println(e.Explain())
		return
	}

	res, err := dice.Roll(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "roll: %v\n", err)
		os.Exit(2)
	}
	Render(os.Stdout, isTTY(os.Stdout), res)
}

// isTTY reports whether f is connected to a terminal. Used to decide whether
// to emit ANSI color codes — pipes and redirects get plain text.
func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
