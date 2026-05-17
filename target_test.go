package dice_test

import (
	"testing"

	"github.com/old-school-gamers/dice"
)

func TestTarget_Hit(t *testing.T) {
	res, err := dice.RollWith("1d20 >= 15", &fixedRoller{values: []int{18}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 18 {
		t.Errorf("Total = %d, want 18", res.Total)
	}
	if res.Target == nil {
		t.Fatal("Target is nil")
	}
	if res.Target.Op != dice.OpGE || res.Target.Value != 15 {
		t.Errorf("Target = %+v, want >=15", res.Target)
	}
	if !res.Hit {
		t.Error("Hit = false, want true")
	}
}

func TestTarget_Miss(t *testing.T) {
	res, err := dice.RollWith("1d20 >= 15", &fixedRoller{values: []int{8}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Hit {
		t.Error("Hit = true, want false")
	}
}

func TestTarget_AllOperators(t *testing.T) {
	cases := []struct {
		expr    string
		roll    int
		wantHit bool
	}{
		{"1d20 >= 15", 15, true},
		{"1d20 >= 15", 14, false},
		{"1d20 <= 5", 5, true},
		{"1d20 <= 5", 6, false},
		{"1d20 > 10", 11, true},
		{"1d20 > 10", 10, false},
		{"1d20 < 10", 9, true},
		{"1d20 < 10", 10, false},
		{"1d20 = 1", 1, true},
		{"1d20 = 1", 2, false},
	}
	for _, tt := range cases {
		t.Run(tt.expr, func(t *testing.T) {
			res, err := dice.RollWith(tt.expr, &fixedRoller{values: []int{tt.roll}})
			if err != nil {
				t.Fatal(err)
			}
			if res.Hit != tt.wantHit {
				t.Errorf("roll %d, want Hit=%v, got %v", tt.roll, tt.wantHit, res.Hit)
			}
		})
	}
}

func TestTarget_NoneByDefault(t *testing.T) {
	res, err := dice.RollWith("1d20", &fixedRoller{values: []int{12}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Target != nil {
		t.Errorf("Target = %+v, want nil", res.Target)
	}
	if res.Hit {
		t.Error("Hit = true, want false (no target)")
	}
}

func TestTarget_AppliesToFinalTotal(t *testing.T) {
	// 1d20 + 5 >= 15: roller 8, total = 13, miss.
	res, err := dice.RollWith("1d20+5 >= 15", &fixedRoller{values: []int{8}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 13 {
		t.Errorf("Total = %d, want 13", res.Total)
	}
	if res.Hit {
		t.Error("Hit = true, want false")
	}
}

func TestTarget_DoesNotAlterTotal(t *testing.T) {
	resWith, err := dice.RollWith("1d20 >= 15", &fixedRoller{values: []int{18}})
	if err != nil {
		t.Fatal(err)
	}
	resWithout, err := dice.RollWith("1d20", &fixedRoller{values: []int{18}})
	if err != nil {
		t.Fatal(err)
	}
	if resWith.Total != resWithout.Total {
		t.Errorf("target altered Total: with=%d without=%d", resWith.Total, resWithout.Total)
	}
}

func TestTarget_ParseErrors(t *testing.T) {
	cases := []string{
		"1d20 >=",        // missing number
		"1d20 >= abc",    // non-number
		"1d20 >= 15 + 1", // extra after target
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
