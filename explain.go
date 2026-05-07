package dice

import (
	"fmt"
	"strings"
)

// Explain returns a plain-English description of what the expression will do.
// It does not roll anything.
func (e *Expression) Explain() string {
	s := e.root.explain()
	if e.target != nil {
		s += "; succeed on " + e.target.ExplainPhrase()
	}
	return s
}

// ExplainPhrase returns a short prose form of the comparison, e.g.
// ">=15" -> "15 or higher".
func (c Comparison) ExplainPhrase() string {
	switch c.Op {
	case OpGE:
		return fmt.Sprintf("%d or higher", c.Value)
	case OpLE:
		return fmt.Sprintf("%d or lower", c.Value)
	case OpGT:
		return fmt.Sprintf("greater than %d", c.Value)
	case OpLT:
		return fmt.Sprintf("less than %d", c.Value)
	case OpEQ:
		return fmt.Sprintf("exactly %d", c.Value)
	}
	return c.String()
}

// --- node explain methods ---

func (n *numNode) explain() string {
	return fmt.Sprintf("%d", n.value)
}

func (n *binOpNode) explain() string {
	// Synthetic unary minus: "-X" was parsed as "0 - X".
	if l, ok := n.left.(*numNode); ok && l.value == 0 && n.op == '-' {
		return "the negative of " + n.right.explain()
	}
	op := "plus"
	if n.op == '-' {
		op = "minus"
	}
	return n.left.explain() + " " + op + " " + n.right.explain()
}

func (n *diceNode) explain() string {
	var sb strings.Builder
	noun := "dice"
	if n.count == 1 {
		noun = "die"
	}
	fmt.Fprintf(&sb, "roll %d %d-sided %s", n.count, n.sides, noun)
	for _, m := range n.mods {
		sb.WriteString(", ")
		sb.WriteString(m.explain())
	}
	return sb.String()
}

func (n *advNode) explain() string {
	prefix := "with advantage, "
	if n.disadvantage {
		prefix = "with disadvantage, "
	}
	return prefix + n.inner.explain()
}

// --- modifier explain methods ---

func (m keepMod) explain() string {
	direction := "highest"
	if !m.highest {
		direction = "lowest"
	}
	if m.n == 1 {
		return "keep the " + direction
	}
	return fmt.Sprintf("keep the %s %d", direction, m.n)
}

func (m dropMod) explain() string {
	direction := "highest"
	if !m.highest {
		direction = "lowest"
	}
	if m.n == 1 {
		return "drop the " + direction
	}
	return fmt.Sprintf("drop the %s %d", direction, m.n)
}

func (m explodeMod) explain() string {
	verb := "explode"
	if m.once {
		verb = "explode once"
	}
	if m.hasCmp {
		return verb + " on " + m.cmp.ExplainPhrase()
	}
	return verb + " on the maximum"
}

func (m rerollMod) explain() string {
	verb := "reroll"
	if m.recursive {
		verb = "reroll recursively"
	}
	return verb + " on " + m.cmp.ExplainPhrase()
}

func (m successMod) explain() string {
	label := "success"
	if m.failure {
		label = "failure"
	}
	return "count each " + m.cmp.ExplainPhrase() + " as a " + label
}
