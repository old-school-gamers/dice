package dice_test

import (
	"testing"

	"github.com/old-school-gamers/dice"
)

func TestKeepHigh(t *testing.T) {
	tests := []struct {
		expr    string
		rolls   []int
		total   int
		dropped []int // values that should be StatusDropped, in roll order
	}{
		{"4d6kh3", []int{6, 2, 5, 1}, 13, []int{1}},
		{"4d6k3", []int{6, 2, 5, 1}, 13, []int{1}}, // 'k' alias for 'kh'
		{"2d20kh1", []int{15, 9}, 15, []int{9}},
		{"2d20kh", []int{9, 15}, 15, []int{9}},     // implicit count of 1
		{"3d6kh1", []int{2, 2, 2}, 2, []int{2, 2}}, // ties: earlier dice win
		{"4d6kh3+2", []int{1, 5, 5, 5}, 17, []int{1}},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			res, err := dice.RollWith(tt.expr, &fixedRoller{values: tt.rolls})
			if err != nil {
				t.Fatalf("RollWith(%q) error: %v", tt.expr, err)
			}
			if res.Total != tt.total {
				t.Errorf("Total = %d, want %d", res.Total, tt.total)
			}
			gotDropped := droppedValues(res)
			if !equalInts(gotDropped, tt.dropped) {
				t.Errorf("dropped = %v, want %v", gotDropped, tt.dropped)
			}
		})
	}
}

func TestKeepLow(t *testing.T) {
	res, err := dice.RollWith("4d6kl1", &fixedRoller{values: []int{6, 2, 5, 1}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 1 {
		t.Errorf("Total = %d, want 1", res.Total)
	}
	if got := droppedValues(res); !equalInts(got, []int{6, 2, 5}) {
		t.Errorf("dropped = %v, want [6 2 5]", got)
	}
}

func TestDropHigh(t *testing.T) {
	res, err := dice.RollWith("4d6dh1", &fixedRoller{values: []int{6, 2, 5, 1}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 8 { // 2+5+1
		t.Errorf("Total = %d, want 8", res.Total)
	}
	if got := droppedValues(res); !equalInts(got, []int{6}) {
		t.Errorf("dropped = %v, want [6]", got)
	}
}

func TestDropLow(t *testing.T) {
	tests := []struct {
		expr  string
		rolls []int
		total int
	}{
		{"4d6dl1", []int{6, 2, 5, 1}, 13}, // standard ability score
		{"4d6dl", []int{6, 2, 5, 1}, 13},  // implicit count of 1
		{"5d6dl2", []int{6, 1, 5, 2, 4}, 15},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			res, err := dice.RollWith(tt.expr, &fixedRoller{values: tt.rolls})
			if err != nil {
				t.Fatal(err)
			}
			if res.Total != tt.total {
				t.Errorf("Total = %d, want %d", res.Total, tt.total)
			}
		})
	}
}

func TestModifierNotationCanonicalized(t *testing.T) {
	res, err := dice.RollWith("4d6kh3", &fixedRoller{values: []int{1, 2, 3, 4}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Groups[0].Notation != "4d6kh3" {
		t.Errorf("Notation = %q, want %q", res.Groups[0].Notation, "4d6kh3")
	}
}

func TestModifierKeepCountClampedToDice(t *testing.T) {
	// kh5 on 3 dice should keep all 3.
	res, err := dice.RollWith("3d6kh5", &fixedRoller{values: []int{2, 4, 6}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 12 {
		t.Errorf("Total = %d, want 12", res.Total)
	}
}

func TestModifierErrors(t *testing.T) {
	cases := []string{
		"1d6kh0", // count must be >= 1
		"1d6kl0",
		"1d6dh0",
		"1d6dl0",
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

// helpers

func droppedValues(res *dice.Result) []int {
	var out []int
	for _, g := range res.Groups {
		for _, d := range g.Dice {
			if d.Status == dice.StatusDropped {
				out = append(out, d.Value)
			}
		}
	}
	return out
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
