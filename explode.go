package dice

import "fmt"

// maxRecursiveRolls caps explosion / reroll chains to keep pathological
// expressions like "1d1x" (every 1 explodes into another 1) from looping
// forever. The cap is per-modifier-application across all chains.
const maxRecursiveRolls = 100

// explodeMod implements x / xo. With no comparison set, the trigger is "value == sides".
type explodeMod struct {
	once   bool
	cmp    Comparison
	hasCmp bool
}

func (m explodeMod) String() string {
	keyword := "x"
	if m.once {
		keyword = "xo"
	}
	if m.hasCmp {
		return keyword + m.cmp.String()
	}
	return keyword
}

func (m explodeMod) trigger(sides int) func(int) bool {
	if m.hasCmp {
		return m.cmp.Match
	}
	return func(v int) bool { return v == sides }
}

func (m explodeMod) apply(d []Die, sides int, r Roller) ([]Die, error) {
	matches := m.trigger(sides)
	rolls := 0

	// Iterate over the original dice only; new dice from explosions are
	// processed within the chain below, not re-scanned here.
	original := len(d)
	for i := 0; i < original; i++ {
		if d[i].Status != StatusKept || !matches(d[i].Value) {
			continue
		}
		// This die triggers a chain.
		d[i].Status = StatusExploded
		for {
			if rolls >= maxRecursiveRolls {
				return nil, fmt.Errorf("dice: explosion cap reached (%d) for %s", maxRecursiveRolls, m.String())
			}
			v, err := r.RollDie(sides)
			if err != nil {
				return nil, err
			}
			rolls++
			d = append(d, Die{Sides: sides, Value: v, Status: StatusKept})
			if m.once || !matches(v) {
				break
			}
			d[len(d)-1].Status = StatusExploded
		}
	}
	return d, nil
}

// rerollMod implements r / rr. The original die is preserved with StatusRerolled
// for display; a replacement die is appended with StatusKept (or StatusRerolled
// itself if recursive and it also matches).
type rerollMod struct {
	recursive bool
	cmp       Comparison
}

func (m rerollMod) String() string {
	keyword := "r"
	if m.recursive {
		keyword = "rr"
	}
	return keyword + m.cmp.String()
}

func (m rerollMod) apply(d []Die, sides int, r Roller) ([]Die, error) {
	rolls := 0
	original := len(d)
	for i := 0; i < original; i++ {
		if d[i].Status != StatusKept || !m.cmp.Match(d[i].Value) {
			continue
		}
		d[i].Status = StatusRerolled
		for {
			if rolls >= maxRecursiveRolls {
				return nil, fmt.Errorf("dice: reroll cap reached (%d) for %s", maxRecursiveRolls, m.String())
			}
			v, err := r.RollDie(sides)
			if err != nil {
				return nil, err
			}
			rolls++
			if !m.recursive || !m.cmp.Match(v) {
				d = append(d, Die{Sides: sides, Value: v, Status: StatusKept})
				break
			}
			d = append(d, Die{Sides: sides, Value: v, Status: StatusRerolled})
		}
	}
	return d, nil
}
