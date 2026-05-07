package dice_test

import (
	"testing"

	"github.com/oldschoolgamers/dice"
)

func TestExplain(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{"1d20", "roll 1 20-sided die"},
		{"2d6", "roll 2 6-sided dice"},
		{"d20", "roll 1 20-sided die"},
		{"1d20+5", "roll 1 20-sided die plus 5"},
		{"2d6+1d4-2", "roll 2 6-sided dice plus roll 1 4-sided die minus 2"},
		{"4d6kh3", "roll 4 6-sided dice, keep the highest 3"},
		{"4d6dl1", "roll 4 6-sided dice, drop the lowest"},
		{"2d20kh", "roll 2 20-sided dice, keep the highest"},
		{"adv(1d20)", "with advantage, roll 1 20-sided die"},
		{"dis(1d20+5)", "with disadvantage, roll 1 20-sided die plus 5"},
		{"1d6x", "roll 1 6-sided die, explode on the maximum"},
		{"1d6x>=5", "roll 1 6-sided die, explode on 5 or higher"},
		{"1d6xo", "roll 1 6-sided die, explode once on the maximum"},
		{"4d6r1", "roll 4 6-sided dice, reroll on exactly 1"},
		{"1d6rr<3", "roll 1 6-sided die, reroll recursively on less than 3"},
		{"5d10cs>=8", "roll 5 10-sided dice, count each 8 or higher as a success"},
		{"5d10cs>=8cf=1", "roll 5 10-sided dice, count each 8 or higher as a success, count each exactly 1 as a failure"},
		{"1d20 >= 15", "roll 1 20-sided die; succeed on 15 or higher"},
		{"1d20 = 20", "roll 1 20-sided die; succeed on exactly 20"},
		{"4d6kh3 >= 15", "roll 4 6-sided dice, keep the highest 3; succeed on 15 or higher"},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			e, err := dice.Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.expr, err)
			}
			got := e.Explain()
			if got != tt.want {
				t.Errorf("Explain() =\n  %q\n  want\n  %q", got, tt.want)
			}
		})
	}
}
