package dice_test

import (
	"testing"

	"github.com/oldschoolgamers/dice"
)

func TestSuccess_Counts(t *testing.T) {
	// 5d10cs>=8: roll 8, 4, 9, 10, 2 → 3 successes.
	res, err := dice.RollWith("5d10cs>=8", &fixedRoller{values: []int{8, 4, 9, 10, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 3 {
		t.Errorf("Total = %d, want 3", res.Total)
	}
	if res.Groups[0].Notation != "5d10cs>=8" {
		t.Errorf("Notation = %q", res.Groups[0].Notation)
	}
	successes := 0
	for _, d := range res.Groups[0].Dice {
		if d.Status == dice.StatusSuccess {
			successes++
		}
	}
	if successes != 3 {
		t.Errorf("success-marked dice = %d, want 3", successes)
	}
}

func TestSuccess_BareNumberMeansGE(t *testing.T) {
	// cs8 should be cs>=8.
	res, err := dice.RollWith("3d10cs8", &fixedRoller{values: []int{7, 8, 9}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 2 {
		t.Errorf("Total = %d, want 2", res.Total)
	}
	if res.Groups[0].Notation != "3d10cs>=8" {
		t.Errorf("Notation = %q, want 3d10cs>=8", res.Groups[0].Notation)
	}
}

func TestFailure_Subtracts(t *testing.T) {
	// 5d10cs>=8cf=1: roll 8, 1, 9, 1, 5 → 2 successes - 2 failures = 0.
	res, err := dice.RollWith("5d10cs>=8cf=1", &fixedRoller{values: []int{8, 1, 9, 1, 5}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 0 {
		t.Errorf("Total = %d, want 0", res.Total)
	}
	successes, failures := 0, 0
	for _, d := range res.Groups[0].Dice {
		switch d.Status {
		case dice.StatusSuccess:
			successes++
		case dice.StatusFailure:
			failures++
		}
	}
	if successes != 2 || failures != 2 {
		t.Errorf("successes=%d failures=%d, want 2/2", successes, failures)
	}
}

func TestFailure_BareNumberMeansLE(t *testing.T) {
	// cf1 should be cf<=1.
	res, err := dice.RollWith("4d6cf1", &fixedRoller{values: []int{1, 2, 1, 6}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != -2 { // two 1s = two failures
		t.Errorf("Total = %d, want -2", res.Total)
	}
	if res.Groups[0].Notation != "4d6cf<=1" {
		t.Errorf("Notation = %q, want 4d6cf<=1", res.Groups[0].Notation)
	}
}

func TestSuccess_AppliesToExplodedChain(t *testing.T) {
	// 1d6xcs>=4: roll 6 (explode) → 4. Both meet >=4. Expect 2 successes total.
	res, err := dice.RollWith("1d6xcs>=4", &fixedRoller{values: []int{6, 4}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 2 {
		t.Errorf("Total = %d, want 2", res.Total)
	}
}

func TestSuccess_WithTargetNumber(t *testing.T) {
	// "5d10 cs>=8 >= 3" — count successes, hit target if at least 3 successes.
	res, err := dice.RollWith("5d10cs>=8 >= 3", &fixedRoller{values: []int{9, 8, 1, 10, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 3 {
		t.Errorf("Total = %d, want 3", res.Total)
	}
	if res.Target == nil || res.Target.Op != dice.OpGE || res.Target.Value != 3 {
		t.Errorf("Target = %+v, want >=3", res.Target)
	}
	if !res.Hit {
		t.Error("Hit = false, want true")
	}
}

func TestSuccess_RequiresComparison(t *testing.T) {
	cases := []string{"1d6cs", "1d6cf"}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			_, err := dice.Parse(expr)
			if err == nil {
				t.Errorf("Parse(%q) = nil error, want error", expr)
			}
		})
	}
}
