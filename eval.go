package dice

import "fmt"

func (n *numNode) eval(_ Roller, _ *Result) (int, error) {
	return n.value, nil
}

func (n *binOpNode) eval(r Roller, res *Result) (int, error) {
	l, err := n.left.eval(r, res)
	if err != nil {
		return 0, err
	}
	rr, err := n.right.eval(r, res)
	if err != nil {
		return 0, err
	}
	switch n.op {
	case '+':
		return l + rr, nil
	case '-':
		return l - rr, nil
	}
	return 0, fmt.Errorf("dice: unknown operator %q", n.op)
}

func (n *diceNode) eval(r Roller, res *Result) (int, error) {
	g := Group{Notation: n.notation}
	for k := 0; k < n.count; k++ {
		v, err := r.RollDie(n.sides)
		if err != nil {
			return 0, err
		}
		if v < 1 || v > n.sides {
			return 0, fmt.Errorf("dice: roller returned %d, out of range [1,%d]", v, n.sides)
		}
		g.Dice = append(g.Dice, Die{Sides: n.sides, Value: v, Status: StatusKept})
	}
	for _, m := range n.mods {
		var err error
		g.Dice, err = m.apply(g.Dice, n.sides, r)
		if err != nil {
			return 0, err
		}
	}
	successCounting := false
	for _, m := range n.mods {
		if _, ok := m.(successMod); ok {
			successCounting = true
			break
		}
	}
	sub := 0
	for _, d := range g.Dice {
		sub += dieValue(d, successCounting)
	}
	g.Subtotal = sub
	res.Groups = append(res.Groups, g)
	return sub, nil
}

// dieValue returns the contribution of a die to its group's subtotal.
// Success counting overrides face value: a Success contributes +1, a Failure -1.
// In success-counting mode, dice that didn't match any threshold contribute 0
// instead of their face value — successes are the only thing that count.
func dieValue(d Die, successCounting bool) int {
	switch d.Status {
	case StatusKept, StatusExploded:
		if successCounting {
			return 0
		}
		return d.Value
	case StatusSuccess:
		return 1
	case StatusFailure:
		return -1
	}
	return 0
}

func (n *advNode) eval(r Roller, res *Result) (int, error) {
	scratchA := &Result{}
	aTotal, err := n.inner.eval(r, scratchA)
	if err != nil {
		return 0, err
	}
	scratchB := &Result{}
	bTotal, err := n.inner.eval(r, scratchB)
	if err != nil {
		return 0, err
	}

	// adv keeps the higher total, dis keeps the lower. Ties favor roll A
	// so behavior is deterministic and matches table convention ("first roll wins ties").
	bWins := bTotal > aTotal
	if n.disadvantage {
		bWins = bTotal < aTotal
	}

	winner, loser, total := scratchA, scratchB, aTotal
	if bWins {
		winner, loser, total = scratchB, scratchA, bTotal
	}

	dice := make([]Die, 0, len(winner.Groups)*2)
	for _, g := range winner.Groups {
		dice = append(dice, g.Dice...)
	}
	for _, g := range loser.Groups {
		for _, d := range g.Dice {
			d.Status = StatusDropped
			dice = append(dice, d)
		}
	}

	res.Groups = append(res.Groups, Group{
		Notation: n.String(),
		Dice:     dice,
		Subtotal: total,
	})
	return total, nil
}
