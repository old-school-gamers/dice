package dice

import "fmt"

// CompOp is a comparison operator used in target numbers and success-counting modifiers.
type CompOp int

const (
	OpGE CompOp = iota // >=
	OpLE               // <=
	OpGT               // >
	OpLT               // <
	OpEQ               // =
)

// Comparison pairs an operator with a threshold value.
type Comparison struct {
	Op    CompOp
	Value int
}

// Match reports whether v satisfies the comparison.
func (c Comparison) Match(v int) bool {
	switch c.Op {
	case OpGE:
		return v >= c.Value
	case OpLE:
		return v <= c.Value
	case OpGT:
		return v > c.Value
	case OpLT:
		return v < c.Value
	case OpEQ:
		return v == c.Value
	}
	return false
}

// String returns the canonical operator+value form, e.g. ">=15".
func (c Comparison) String() string {
	return c.Op.String() + fmt.Sprintf("%d", c.Value)
}

// String returns the operator literal.
func (op CompOp) String() string {
	switch op {
	case OpGE:
		return ">="
	case OpLE:
		return "<="
	case OpGT:
		return ">"
	case OpLT:
		return "<"
	case OpEQ:
		return "="
	}
	return "?"
}

// parseComparison consumes a comparison at the parser's current position.
// It expects either an explicit operator (>=/<=/>/</=) followed by a number,
// or a bare number (in which case the caller's defaultOp is used — for
// modifiers like 'x4' meaning "explode on >=4" or "explode on =4" depending
// on the modifier's convention).
func (p *parser) parseComparison(defaultOp CompOp, allowBareNumber bool) (Comparison, bool, error) {
	t := p.peek()
	op := defaultOp
	hasOp := false
	switch t.kind {
	case tokGE:
		op, hasOp = OpGE, true
		p.advance()
	case tokLE:
		op, hasOp = OpLE, true
		p.advance()
	case tokGT:
		op, hasOp = OpGT, true
		p.advance()
	case tokLT:
		op, hasOp = OpLT, true
		p.advance()
	case tokEQ:
		op, hasOp = OpEQ, true
		p.advance()
	}
	nt := p.peek()
	if nt.kind != tokNum {
		if !hasOp {
			return Comparison{}, false, nil
		}
		return Comparison{}, false, fmt.Errorf("dice: expected number after comparison operator at position %d", nt.pos)
	}
	if !hasOp && !allowBareNumber {
		return Comparison{}, false, nil
	}
	p.advance()
	v, err := parseInt(nt.text)
	if err != nil {
		return Comparison{}, false, fmt.Errorf("dice: invalid number %q: %w", nt.text, err)
	}
	return Comparison{Op: op, Value: v}, true, nil
}

func parseInt(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, err
	}
	return n, nil
}
