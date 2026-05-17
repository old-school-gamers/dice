package dice_test

import (
	"strings"
	"testing"

	"github.com/old-school-gamers/dice"
)

func TestExplode_OnMax(t *testing.T) {
	// 1d6x: roll 6 → explode → roll 4 → stop. Total = 6+4 = 10.
	res, err := dice.RollWith("1d6x", &fixedRoller{values: []int{6, 4}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 10 {
		t.Errorf("Total = %d, want 10", res.Total)
	}
	g := res.Groups[0]
	if g.Notation != "1d6x" {
		t.Errorf("Notation = %q, want 1d6x", g.Notation)
	}
	if len(g.Dice) != 2 {
		t.Fatalf("got %d dice, want 2", len(g.Dice))
	}
	if g.Dice[0].Value != 6 || g.Dice[0].Status != dice.StatusExploded {
		t.Errorf("die[0] = %+v, want value=6 exploded", g.Dice[0])
	}
	if g.Dice[1].Value != 4 || g.Dice[1].Status != dice.StatusKept {
		t.Errorf("die[1] = %+v, want value=4 kept", g.Dice[1])
	}
}

func TestExplode_Chain(t *testing.T) {
	// 1d6x: 6 → 6 → 6 → 2. Total = 20.
	res, err := dice.RollWith("1d6x", &fixedRoller{values: []int{6, 6, 6, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 20 {
		t.Errorf("Total = %d, want 20", res.Total)
	}
	g := res.Groups[0]
	if len(g.Dice) != 4 {
		t.Fatalf("got %d dice, want 4", len(g.Dice))
	}
	wantStatus := []dice.Status{
		dice.StatusExploded, dice.StatusExploded, dice.StatusExploded, dice.StatusKept,
	}
	for i, d := range g.Dice {
		if d.Status != wantStatus[i] {
			t.Errorf("die[%d] status = %v, want %v", i, d.Status, wantStatus[i])
		}
	}
}

func TestExplode_Once(t *testing.T) {
	// 1d6xo: 6 → 6 → stop (once only). Total = 12; second 6 should be Kept, not Exploded.
	res, err := dice.RollWith("1d6xo", &fixedRoller{values: []int{6, 6}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 12 {
		t.Errorf("Total = %d, want 12", res.Total)
	}
	g := res.Groups[0]
	if len(g.Dice) != 2 {
		t.Fatalf("got %d dice, want 2", len(g.Dice))
	}
	if g.Dice[0].Status != dice.StatusExploded {
		t.Errorf("die[0] should be exploded")
	}
	if g.Dice[1].Status != dice.StatusKept {
		t.Errorf("die[1] should be kept (not exploded again)")
	}
}

func TestExplode_BareNumberMeansGE(t *testing.T) {
	// 1d10x8: explode on >=8. Roll 9 → explode → 7 → stop. Total = 16.
	res, err := dice.RollWith("1d10x8", &fixedRoller{values: []int{9, 7}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 16 {
		t.Errorf("Total = %d, want 16", res.Total)
	}
	if res.Groups[0].Notation != "1d10x>=8" {
		t.Errorf("Notation = %q, want 1d10x>=8", res.Groups[0].Notation)
	}
}

func TestExplode_ExplicitComparison(t *testing.T) {
	// 1d6x>=5: 5 → 3. Total 8.
	res, err := dice.RollWith("1d6x>=5", &fixedRoller{values: []int{5, 3}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 8 {
		t.Errorf("Total = %d, want 8", res.Total)
	}
}

func TestExplode_DoesNotTriggerOnNonMatch(t *testing.T) {
	// 1d6x: 3 (no explosion). Total = 3.
	res, err := dice.RollWith("1d6x", &fixedRoller{values: []int{3}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 3 {
		t.Errorf("Total = %d, want 3", res.Total)
	}
	if len(res.Groups[0].Dice) != 1 {
		t.Errorf("got %d dice, want 1", len(res.Groups[0].Dice))
	}
}

func TestExplode_RunawayCapped(t *testing.T) {
	// 1d1x always rolls 1 and 1==max so explodes forever — should hit the cap and error.
	values := make([]int, 200)
	for i := range values {
		values[i] = 1
	}
	_, err := dice.RollWith("1d1x", &fixedRoller{values: values})
	if err == nil {
		t.Fatal("expected explosion cap error, got nil")
	}
	if !strings.Contains(err.Error(), "explosion cap") {
		t.Errorf("err = %v, want 'explosion cap'", err)
	}
}

func TestReroll_Once(t *testing.T) {
	// 4d6r1: any 1 is rerolled once. Rolls: 1, 4, 5, 1. Rerolls: 3, 6.
	// Final dice in result: orig 1 (rerolled), 4, 5, orig 1 (rerolled), reroll 3, reroll 6.
	res, err := dice.RollWith("4d6r1", &fixedRoller{values: []int{1, 4, 5, 1, 3, 6}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 4+5+3+6 {
		t.Errorf("Total = %d, want 18", res.Total)
	}
	g := res.Groups[0]
	if g.Notation != "4d6r=1" {
		t.Errorf("Notation = %q, want 4d6r=1", g.Notation)
	}
	if len(g.Dice) != 6 {
		t.Fatalf("got %d dice, want 6", len(g.Dice))
	}
	// Original 1s should be StatusRerolled.
	rerolled := 0
	for _, d := range g.Dice {
		if d.Status == dice.StatusRerolled {
			rerolled++
		}
	}
	if rerolled != 2 {
		t.Errorf("rerolled count = %d, want 2", rerolled)
	}
}

func TestReroll_OnceDoesNotChain(t *testing.T) {
	// 1d6r<3: 1 (matches) → reroll 2 (also matches but only one reroll allowed). Total = 2.
	res, err := dice.RollWith("1d6r<3", &fixedRoller{values: []int{1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 2 {
		t.Errorf("Total = %d, want 2", res.Total)
	}
	if len(res.Groups[0].Dice) != 2 {
		t.Errorf("got %d dice, want 2", len(res.Groups[0].Dice))
	}
}

func TestReroll_Recursive(t *testing.T) {
	// 1d6rr<3: 1 → 2 → 5. Final value 5; first two rerolled.
	res, err := dice.RollWith("1d6rr<3", &fixedRoller{values: []int{1, 2, 5}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 5 {
		t.Errorf("Total = %d, want 5", res.Total)
	}
	g := res.Groups[0]
	if len(g.Dice) != 3 {
		t.Fatalf("got %d dice, want 3", len(g.Dice))
	}
	wantStatus := []dice.Status{dice.StatusRerolled, dice.StatusRerolled, dice.StatusKept}
	for i, d := range g.Dice {
		if d.Status != wantStatus[i] {
			t.Errorf("die[%d] status = %v, want %v", i, d.Status, wantStatus[i])
		}
	}
}

func TestReroll_NoMatchNoChange(t *testing.T) {
	res, err := dice.RollWith("1d6r1", &fixedRoller{values: []int{4}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 4 {
		t.Errorf("Total = %d, want 4", res.Total)
	}
}

func TestReroll_RunawayCapped(t *testing.T) {
	values := make([]int, 200)
	for i := range values {
		values[i] = 1
	}
	_, err := dice.RollWith("1d6rr<6", &fixedRoller{values: values})
	if err == nil {
		t.Fatal("expected reroll cap error, got nil")
	}
	if !strings.Contains(err.Error(), "reroll cap") {
		t.Errorf("err = %v, want 'reroll cap'", err)
	}
}

func TestReroll_RequiresComparison(t *testing.T) {
	cases := []string{"1d6r", "1d6rr"}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			_, err := dice.Parse(expr)
			if err == nil {
				t.Errorf("Parse(%q) = nil error, want error", expr)
			}
		})
	}
}

func TestExplodeAndRerollCompose(t *testing.T) {
	// 4d6r1x: reroll 1s (once), then explode on 6.
	// Rolls: 1, 6, 3, 5. Reroll the 1: 2. Explode the 6: → 4.
	// Order in result: orig 1 (rerolled), 6 (exploded), 3, 5, reroll 2, explosion 4.
	res, err := dice.RollWith("4d6r1x", &fixedRoller{values: []int{1, 6, 3, 5, 2, 4}})
	if err != nil {
		t.Fatal(err)
	}
	// Total = 6 + 3 + 5 + 2 + 4 = 20 (orig 1 dropped via reroll).
	if res.Total != 20 {
		t.Errorf("Total = %d, want 20", res.Total)
	}
}
