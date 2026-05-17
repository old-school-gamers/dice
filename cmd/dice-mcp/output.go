package main

import (
	"fmt"
	"strings"

	"github.com/old-school-gamers/dice"
)

// RollOutput is the structured response returned for a successful roll tool call.
type RollOutput struct {
	Expression string        `json:"expression"`
	Total      int           `json:"total"`
	Target     *TargetOutput `json:"target,omitempty"`
	Hit        *bool         `json:"hit,omitempty"`
	Groups     []GroupOutput `json:"groups"`
}

// TargetOutput mirrors dice.Comparison in JSON-friendly form.
type TargetOutput struct {
	Op    string `json:"op"`
	Value int    `json:"value"`
}

// GroupOutput is one dice expression's per-die breakdown.
type GroupOutput struct {
	Notation string      `json:"notation"`
	Subtotal int         `json:"subtotal"`
	Dice     []DieOutput `json:"dice"`
}

// DieOutput is a single rolled die.
type DieOutput struct {
	Sides  int    `json:"sides"`
	Value  int    `json:"value"`
	Status string `json:"status"`
}

// ExplainOutput is the structured response returned by the explain tool.
type ExplainOutput struct {
	Expression  string `json:"expression"`
	Description string `json:"description"`
}

// buildRollOutput converts a dice.Result into the wire format.
func buildRollOutput(res *dice.Result) RollOutput {
	out := RollOutput{
		Expression: res.Expression,
		Total:      res.Total,
		Groups:     make([]GroupOutput, 0, len(res.Groups)),
	}
	if res.Target != nil {
		out.Target = &TargetOutput{Op: res.Target.Op.String(), Value: res.Target.Value}
		hit := res.Hit
		out.Hit = &hit
	}
	for _, g := range res.Groups {
		go_ := GroupOutput{Notation: g.Notation, Subtotal: g.Subtotal, Dice: make([]DieOutput, len(g.Dice))}
		for i, d := range g.Dice {
			go_.Dice[i] = DieOutput{Sides: d.Sides, Value: d.Value, Status: statusString(d.Status)}
		}
		out.Groups = append(out.Groups, go_)
	}
	return out
}

// formatRollText is the human-readable summary that accompanies the
// structured output. Kept short — the structured payload carries detail.
func formatRollText(res *dice.Result) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s = %d", res.Expression, res.Total)
	if res.Target != nil {
		if res.Hit {
			sb.WriteString(" (HIT)")
		} else {
			sb.WriteString(" (MISS)")
		}
	}
	if len(res.Groups) > 0 {
		sb.WriteString("\n")
		for i, g := range res.Groups {
			if i > 0 {
				sb.WriteString("\n")
			}
			fmt.Fprintf(&sb, "  %s:", g.Notation)
			for _, d := range g.Dice {
				fmt.Fprintf(&sb, " %d", d.Value)
				if mark := statusMark(d.Status); mark != "" {
					sb.WriteString(mark)
				}
			}
		}
	}
	return sb.String()
}

func statusString(s dice.Status) string {
	switch s {
	case dice.StatusKept:
		return "kept"
	case dice.StatusDropped:
		return "dropped"
	case dice.StatusExploded:
		return "exploded"
	case dice.StatusRerolled:
		return "rerolled"
	case dice.StatusSuccess:
		return "success"
	case dice.StatusFailure:
		return "failure"
	}
	return "unknown"
}

func statusMark(s dice.Status) string {
	switch s {
	case dice.StatusDropped:
		return "(dropped)"
	case dice.StatusRerolled:
		return "(rerolled)"
	case dice.StatusExploded:
		return "!"
	case dice.StatusSuccess:
		return "*"
	case dice.StatusFailure:
		return "x"
	}
	return ""
}
