package dice_test

import (
	"testing"

	"github.com/old-school-gamers/dice"
)

func TestAdvantage(t *testing.T) {
	r := &fixedRoller{values: []int{9, 15}}
	res, err := dice.RollWith("adv(1d20)", r)
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 15 {
		t.Errorf("Total = %d, want 15", res.Total)
	}
	if len(res.Groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(res.Groups))
	}
	g := res.Groups[0]
	if g.Notation != "adv(1d20)" {
		t.Errorf("Notation = %q, want adv(1d20)", g.Notation)
	}
	if len(g.Dice) != 2 {
		t.Fatalf("dice count = %d, want 2", len(g.Dice))
	}
	// Winner came from the second roll (15), so dice order is winner-first then loser-dropped.
	if g.Dice[0].Value != 15 || g.Dice[0].Status != dice.StatusKept {
		t.Errorf("winner die = %+v", g.Dice[0])
	}
	if g.Dice[1].Value != 9 || g.Dice[1].Status != dice.StatusDropped {
		t.Errorf("loser die = %+v", g.Dice[1])
	}
}

func TestDisadvantage(t *testing.T) {
	r := &fixedRoller{values: []int{9, 15}}
	res, err := dice.RollWith("dis(1d20)", r)
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 9 {
		t.Errorf("Total = %d, want 9", res.Total)
	}
	g := res.Groups[0]
	if g.Notation != "dis(1d20)" {
		t.Errorf("Notation = %q, want dis(1d20)", g.Notation)
	}
	if g.Dice[0].Value != 9 || g.Dice[0].Status != dice.StatusKept {
		t.Errorf("winner die = %+v", g.Dice[0])
	}
	if g.Dice[1].Value != 15 || g.Dice[1].Status != dice.StatusDropped {
		t.Errorf("loser die = %+v", g.Dice[1])
	}
}

func TestAdvantageTiesFavorFirst(t *testing.T) {
	r := &fixedRoller{values: []int{12, 12}}
	res, err := dice.RollWith("adv(1d20)", r)
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 12 {
		t.Errorf("Total = %d, want 12", res.Total)
	}
	// Roll A wins ties; second roll's dice should be marked dropped.
	g := res.Groups[0]
	if g.Dice[0].Status != dice.StatusKept || g.Dice[1].Status != dice.StatusDropped {
		t.Errorf("tie behavior wrong: %+v", g.Dice)
	}
}

func TestAdvantageWithModifierAroundIt(t *testing.T) {
	r := &fixedRoller{values: []int{8, 16}}
	res, err := dice.RollWith("adv(1d20)+5", r)
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 21 { // 16 + 5
		t.Errorf("Total = %d, want 21", res.Total)
	}
}

func TestAdvantageOnCompoundExpression(t *testing.T) {
	// adv(2d6+1) — 2d6+1 rolled twice; higher total wins.
	// Roll A: 1+2+1 = 4, Roll B: 5+6+1 = 12. Winner: B = 12.
	r := &fixedRoller{values: []int{1, 2, 5, 6}}
	res, err := dice.RollWith("adv(2d6+1)", r)
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 12 {
		t.Errorf("Total = %d, want 12", res.Total)
	}
	g := res.Groups[0]
	if g.Notation != "adv(2d6+1)" {
		t.Errorf("Notation = %q, want adv(2d6+1)", g.Notation)
	}
	// Winner first (5,6) kept, then loser (1,2) dropped.
	if len(g.Dice) != 4 {
		t.Fatalf("dice count = %d, want 4", len(g.Dice))
	}
	wantValues := []int{5, 6, 1, 2}
	wantStatus := []dice.Status{dice.StatusKept, dice.StatusKept, dice.StatusDropped, dice.StatusDropped}
	for i, d := range g.Dice {
		if d.Value != wantValues[i] || d.Status != wantStatus[i] {
			t.Errorf("die[%d] = %+v, want value=%d status=%v", i, d, wantValues[i], wantStatus[i])
		}
	}
}

func TestAdvantageErrors(t *testing.T) {
	cases := []string{
		"adv",      // no parens
		"adv(",     // no inner, no closing
		"adv()",    // empty inner
		"adv(1d20", // unclosed
		"dis(1d20", // unclosed
		"adv 1d20", // no parens at all
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
