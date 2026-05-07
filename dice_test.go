package dice_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/oldschoolgamers/dice"
)

// fixedRoller returns predetermined die values in order. It panics in tests
// when more rolls are requested than provided so missing setup is loud.
type fixedRoller struct {
	values []int
	i      int
}

func (r *fixedRoller) RollDie(sides int) (int, error) {
	if r.i >= len(r.values) {
		return 0, fmt.Errorf("fixedRoller: out of values (asked for d%d, used %d)", sides, r.i)
	}
	v := r.values[r.i]
	r.i++
	return v, nil
}

func TestRoll_BasicAndArithmetic(t *testing.T) {
	tests := []struct {
		expr  string
		rolls []int
		total int
	}{
		{"1d6", []int{4}, 4},
		{"2d6", []int{3, 5}, 8},
		{"d20", []int{15}, 15},
		{"D20", []int{15}, 15}, // case-insensitive
		{"2d6+3", []int{2, 4}, 9},
		{"1d20-1", []int{15}, 14},
		{"2d6 + 1d4", []int{3, 4, 2}, 9},
		{"4", nil, 4},
		{"2+3", nil, 5},
		{"10-4", nil, 6},
		{"(1d6)+1", []int{4}, 5},
		{"2*0+1d6", nil, 0}, // sentinel: not parsed (would fail) — see TestRoll_Errors
		{"1d20+5+1d4", []int{12, 2}, 19},
		{"-1d6+10", []int{4}, 6},
		{"0d6", nil, 0},
		{"1d1", nil, 1}, // pinned by roller, but 1d1 is deterministic anyway
	}
	for _, tt := range tests {
		if strings.Contains(tt.expr, "*") {
			continue // covered in TestRoll_Errors
		}
		t.Run(tt.expr, func(t *testing.T) {
			rolls := tt.rolls
			if tt.expr == "1d1" {
				rolls = []int{1}
			}
			r := &fixedRoller{values: rolls}
			res, err := dice.RollWith(tt.expr, r)
			if err != nil {
				t.Fatalf("RollWith(%q) error: %v", tt.expr, err)
			}
			if res.Total != tt.total {
				t.Errorf("Total = %d, want %d", res.Total, tt.total)
			}
			if r.i != len(rolls) {
				t.Errorf("used %d of %d rolls", r.i, len(rolls))
			}
		})
	}
}

func TestRoll_GroupsRecorded(t *testing.T) {
	r := &fixedRoller{values: []int{3, 5}}
	res, err := dice.RollWith("2d6+1", r)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(res.Groups))
	}
	g := res.Groups[0]
	if g.Notation != "2d6" {
		t.Errorf("Notation = %q, want %q", g.Notation, "2d6")
	}
	if g.Subtotal != 8 {
		t.Errorf("Subtotal = %d, want 8", g.Subtotal)
	}
	if len(g.Dice) != 2 || g.Dice[0].Value != 3 || g.Dice[1].Value != 5 {
		t.Errorf("dice = %+v", g.Dice)
	}
	for _, d := range g.Dice {
		if d.Status != dice.StatusKept {
			t.Errorf("die %+v not kept", d)
		}
		if d.Sides != 6 {
			t.Errorf("die sides = %d, want 6", d.Sides)
		}
	}
}

func TestRoll_PureNumberHasNoGroups(t *testing.T) {
	res, err := dice.RollWith("7", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Groups) != 0 {
		t.Errorf("expected 0 groups for pure number, got %d", len(res.Groups))
	}
	if res.Total != 7 {
		t.Errorf("Total = %d, want 7", res.Total)
	}
}

func TestRoll_Errors(t *testing.T) {
	cases := []string{
		"",        // empty
		"   ",     // whitespace only
		"abc",     // gibberish
		"1d",      // missing sides
		"d",       // dangling 'd' with no sides
		"1d0",     // zero sides
		"(1d6",    // unclosed paren
		"1d6)",    // stray closing paren
		"1+",      // trailing operator
		"2*3",     // unsupported operator
		"1d6 1d6", // two terms with no operator
	}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			_, err := dice.Parse(expr)
			if err == nil {
				t.Errorf("Parse(%q) = nil error, want error", expr)
			}
		})
	}
}

func TestRoll_DefaultRollerProducesValuesInRange(t *testing.T) {
	// Smoke test the crypto/rand path; we just want non-error and bounds.
	for range 50 {
		res, err := dice.Roll("1d20")
		if err != nil {
			t.Fatalf("Roll error: %v", err)
		}
		if res.Total < 1 || res.Total > 20 {
			t.Errorf("Total = %d, want in [1,20]", res.Total)
		}
	}
}

func TestRoll_RollerErrorPropagates(t *testing.T) {
	bad := errRoller{err: errors.New("boom")}
	_, err := dice.RollWith("1d6", bad)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("err = %v, want to contain 'boom'", err)
	}
}

type errRoller struct{ err error }

func (e errRoller) RollDie(int) (int, error) { return 0, e.err }
