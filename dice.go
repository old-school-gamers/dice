package dice

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Status describes the role of a die in the final result.
type Status int

const (
	// StatusKept is the default: the die contributed to its group's subtotal.
	StatusKept Status = iota
	// StatusDropped means the die was rolled but excluded by a keep/drop modifier
	// or by adv()/dis() selection. Its Value is preserved so consumers can render it.
	StatusDropped
	// StatusExploded marks a die that triggered an explosion (a follow-up die
	// was rolled because of it). The exploded die still counts toward the total.
	StatusExploded
	// StatusRerolled marks a die whose value was thrown away by a reroll modifier.
	// Its Value is preserved for display.
	StatusRerolled
	// StatusSuccess marks a die that counted as a success in success-counting mode.
	StatusSuccess
	// StatusFailure marks a die that counted as a failure in success-counting mode.
	StatusFailure
)

// Die is a single die roll within a Result.
type Die struct {
	Sides  int
	Value  int
	Status Status
}

// Group is one dice expression's contribution to a Result.
// A bare number literal does not produce a group.
type Group struct {
	Notation string // e.g. "2d6kh1"
	Dice     []Die
	Subtotal int // contribution of this group to Result.Total
}

// Result is the full result of evaluating a dice expression.
type Result struct {
	Expression string
	Groups     []Group
	Total      int
	// Target is the expression-level target number (e.g. ">=15"), if the
	// caller specified one. It does not affect Total — it's metadata for UI
	// highlighting (green on hit, red on miss).
	Target *Comparison
	// Hit is true when Target is non-nil and Total satisfies it.
	Hit bool
}

// Roller produces individual die rolls. Implementations must return values
// in [1, sides].
type Roller interface {
	RollDie(sides int) (int, error)
}

// Roll parses and rolls a dice expression using crypto/rand.
func Roll(expr string) (*Result, error) {
	return RollWith(expr, defaultRoller{})
}

// RollWith parses and rolls a dice expression using the provided Roller.
// A nil Roller falls back to the crypto/rand default.
func RollWith(expr string, r Roller) (*Result, error) {
	e, err := Parse(expr)
	if err != nil {
		return nil, err
	}
	return e.Roll(r)
}

// Expression is a parsed dice expression, ready to roll repeatedly.
type Expression struct {
	src    string
	root   node
	target *Comparison // optional trailing target like ">=15"
}

// Parse parses a dice expression without rolling it.
func Parse(s string) (*Expression, error) {
	root, target, err := parse(s)
	if err != nil {
		return nil, err
	}
	return &Expression{src: s, root: root, target: target}, nil
}

// Source returns the original expression text.
func (e *Expression) Source() string { return e.src }

// Target returns the expression-level target number, or nil if none.
func (e *Expression) Target() *Comparison {
	if e.target == nil {
		return nil
	}
	c := *e.target
	return &c
}

// Roll evaluates the parsed expression. A nil Roller uses crypto/rand.
func (e *Expression) Roll(r Roller) (*Result, error) {
	if r == nil {
		r = defaultRoller{}
	}
	res := &Result{Expression: e.src}
	total, err := e.root.eval(r, res)
	if err != nil {
		return nil, err
	}
	res.Total = total
	if e.target != nil {
		t := *e.target
		res.Target = &t
		res.Hit = t.Match(total)
	}
	return res, nil
}

// defaultRoller draws from crypto/rand. rand.Int rejects biased samples,
// so all values in [1, sides] are uniform.
type defaultRoller struct{}

func (defaultRoller) RollDie(sides int) (int, error) {
	if sides < 1 {
		return 0, fmt.Errorf("dice: invalid die size %d", sides)
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 1, nil
}
