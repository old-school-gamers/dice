package dice

import (
	"fmt"
	"sort"
	"strings"
)

// modifier transforms a slice of rolled dice. It may mutate Status fields
// and append additional dice (exploding, rerolls). Implementations must be
// deterministic given identical inputs.
type modifier interface {
	apply(d []Die, sides int, r Roller) ([]Die, error)
	String() string  // canonical suffix, e.g. "kh3"
	explain() string // prose description used by Expression.Explain
}

// keptIndices returns indices of dice currently marked StatusKept,
// in original (roll) order.
func keptIndices(d []Die) []int {
	out := make([]int, 0, len(d))
	for i, x := range d {
		if x.Status == StatusKept {
			out = append(out, i)
		}
	}
	return out
}

// pickByValue returns the n indices from src whose Die.Value satisfies
// the ranking direction. highest=true keeps largest values, false keeps
// smallest. Ties are broken by earlier roll order.
func pickByValue(d []Die, src []int, n int, highest bool) []int {
	if n >= len(src) {
		return src
	}
	idx := make([]int, len(src))
	copy(idx, src)
	sort.SliceStable(idx, func(a, b int) bool {
		va, vb := d[idx[a]].Value, d[idx[b]].Value
		if highest {
			return va > vb
		}
		return va < vb
	})
	return idx[:n]
}

func dropOthers(d []Die, kept []int) {
	keep := make(map[int]bool, len(kept))
	for _, i := range kept {
		keep[i] = true
	}
	for i := range d {
		if d[i].Status == StatusKept && !keep[i] {
			d[i].Status = StatusDropped
		}
	}
}

// keepMod implements kh/kl by keeping N currently-kept dice and dropping the rest.
type keepMod struct {
	n        int
	highest  bool
	notation string
}

func (m keepMod) apply(d []Die, _ int, _ Roller) ([]Die, error) {
	live := keptIndices(d)
	winners := pickByValue(d, live, m.n, m.highest)
	dropOthers(d, winners)
	return d, nil
}

func (m keepMod) String() string { return m.notation }

// dropMod implements dh/dl by dropping N currently-kept dice.
type dropMod struct {
	n        int
	highest  bool
	notation string
}

func (m dropMod) apply(d []Die, _ int, _ Roller) ([]Die, error) {
	live := keptIndices(d)
	losers := pickByValue(d, live, m.n, m.highest)
	loseSet := make(map[int]bool, len(losers))
	for _, i := range losers {
		loseSet[i] = true
	}
	for i := range d {
		if d[i].Status == StatusKept && loseSet[i] {
			d[i].Status = StatusDropped
		}
	}
	return d, nil
}

func (m dropMod) String() string { return m.notation }

// modifierKeywords lists known modifier identifiers, longest first so that
// a shared IDENT token like "xcs" or "khdl" is split correctly (kh before k,
// xo before x, rr before r, etc.).
var modifierKeywords = []string{"kh", "kl", "dh", "dl", "xo", "rr", "cs", "cf", "k", "x", "r"}

// parseModifier parses a single modifier. Returns ok=false (and no error)
// when the next token is not a recognized modifier — that signals the
// modifier loop to stop without consuming anything.
//
// When the current IDENT token starts with a modifier keyword followed by
// more letters (e.g. "xcs"), the keyword prefix is consumed and the remaining
// letters are pushed back into the token stream as a fresh IDENT for the next
// iteration.
func (p *parser) parseModifier() (modifier, bool, error) {
	t := p.peek()
	if t.kind != tokIdent {
		return nil, false, nil
	}
	var kw string
	for _, k := range modifierKeywords {
		if strings.HasPrefix(t.text, k) {
			kw = k
			break
		}
	}
	if kw == "" {
		return nil, false, nil
	}
	if len(t.text) > len(kw) {
		// Rewrite this token to drop the consumed prefix; do not advance.
		p.toks[p.i] = token{kind: tokIdent, text: t.text[len(kw):], pos: t.pos + len(kw)}
	} else {
		p.advance()
	}
	return p.parseModifierByKeyword(kw)
}

func (p *parser) parseModifierByKeyword(kw string) (modifier, bool, error) {
	switch kw {
	case "k", "kh":
		n, suffix, err := p.parseModCount(kw, 1)
		if err != nil {
			return nil, false, err
		}
		return keepMod{n: n, highest: true, notation: suffix}, true, nil
	case "kl":
		n, suffix, err := p.parseModCount(kw, 1)
		if err != nil {
			return nil, false, err
		}
		return keepMod{n: n, highest: false, notation: suffix}, true, nil
	case "dh":
		n, suffix, err := p.parseModCount(kw, 1)
		if err != nil {
			return nil, false, err
		}
		return dropMod{n: n, highest: true, notation: suffix}, true, nil
	case "dl":
		n, suffix, err := p.parseModCount(kw, 1)
		if err != nil {
			return nil, false, err
		}
		return dropMod{n: n, highest: false, notation: suffix}, true, nil
	case "x", "xo":
		// Bare number means ">=N" (Foundry/Roll20 convention).
		cmp, hasCmp, err := p.parseComparison(OpGE, true)
		if err != nil {
			return nil, false, err
		}
		return explodeMod{once: kw == "xo", cmp: cmp, hasCmp: hasCmp}, true, nil
	case "r", "rr":
		// Bare number means "=N" (e.g. "r1" = reroll 1s).
		cmp, hasCmp, err := p.parseComparison(OpEQ, true)
		if err != nil {
			return nil, false, err
		}
		if !hasCmp {
			return nil, false, fmt.Errorf("dice: %s requires a comparison at position %d", kw, p.peek().pos)
		}
		return rerollMod{recursive: kw == "rr", cmp: cmp}, true, nil
	case "cs", "cf":
		failure := kw == "cf"
		// Bare number means ">=N" for cs (typical), "<=N" for cf (botches).
		def := OpGE
		if failure {
			def = OpLE
		}
		cmp, hasCmp, err := p.parseComparison(def, true)
		if err != nil {
			return nil, false, err
		}
		if !hasCmp {
			return nil, false, fmt.Errorf("dice: %s requires a comparison at position %d", kw, p.peek().pos)
		}
		return successMod{failure: failure, cmp: cmp}, true, nil
	}
	return nil, false, nil
}

// parseModCount reads an optional integer following a modifier keyword.
// def is the value used when no number is present.
func (p *parser) parseModCount(keyword string, def int) (int, string, error) {
	if p.peek().kind != tokNum {
		return def, fmt.Sprintf("%s%d", keyword, def), nil
	}
	tok := p.advance()
	n, err := parsePositive(tok.text)
	if err != nil {
		return 0, "", fmt.Errorf("dice: %s expects positive integer at position %d: %w", keyword, tok.pos, err)
	}
	return n, fmt.Sprintf("%s%d", keyword, n), nil
}

func parsePositive(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, fmt.Errorf("must be >= 1, got %d", n)
	}
	return n, nil
}
