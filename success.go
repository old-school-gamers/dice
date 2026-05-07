package dice

// successMod implements cs (count successes) and cf (count failures).
// Eligible dice are those with face values still in play (Kept or Exploded).
// Marked dice contribute +1 (success) or -1 (failure) to the group's Subtotal,
// replacing their face-value contribution.
type successMod struct {
	failure bool // false = cs, true = cf
	cmp     Comparison
}

func (m successMod) String() string {
	keyword := "cs"
	if m.failure {
		keyword = "cf"
	}
	return keyword + m.cmp.String()
}

func (m successMod) apply(d []Die, _ int, _ Roller) ([]Die, error) {
	target := StatusSuccess
	if m.failure {
		target = StatusFailure
	}
	for i := range d {
		if d[i].Status != StatusKept && d[i].Status != StatusExploded {
			continue
		}
		if m.cmp.Match(d[i].Value) {
			d[i].Status = target
		}
	}
	return d, nil
}
