package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/old-school-gamers/dice"
)

type fixedRoller struct {
	values []int
	i      int
}

func (r *fixedRoller) RollDie(int) (int, error) {
	if r.i >= len(r.values) {
		return 0, fmt.Errorf("fixedRoller: out of values (used %d)", r.i)
	}
	v := r.values[r.i]
	r.i++
	return v, nil
}

func TestHandleRoll_Basic(t *testing.T) {
	h := &handlers{roller: &fixedRoller{values: []int{4, 5}}}
	res, structured, err := h.handleRoll(context.Background(), nil, RollInput{Expression: "2d6+3"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError: %+v", res)
	}
	out, ok := structured.(RollOutput)
	if !ok {
		t.Fatalf("structured = %T, want RollOutput", structured)
	}
	if out.Total != 12 {
		t.Errorf("Total = %d, want 12", out.Total)
	}
	if len(out.Groups) != 1 || out.Groups[0].Notation != "2d6" {
		t.Errorf("Groups = %+v", out.Groups)
	}
}

func TestHandleRoll_TargetHit(t *testing.T) {
	h := &handlers{roller: &fixedRoller{values: []int{18}}}
	_, structured, err := h.handleRoll(context.Background(), nil, RollInput{Expression: "1d20 >= 15"})
	if err != nil {
		t.Fatal(err)
	}
	out := structured.(RollOutput)
	if out.Target == nil || out.Target.Op != ">=" || out.Target.Value != 15 {
		t.Errorf("Target = %+v", out.Target)
	}
	if out.Hit == nil || !*out.Hit {
		t.Errorf("Hit = %v, want true", out.Hit)
	}
}

func TestHandleRoll_TargetMiss(t *testing.T) {
	h := &handlers{roller: &fixedRoller{values: []int{4}}}
	_, structured, err := h.handleRoll(context.Background(), nil, RollInput{Expression: "1d20 >= 15"})
	if err != nil {
		t.Fatal(err)
	}
	out := structured.(RollOutput)
	if out.Hit == nil || *out.Hit {
		t.Errorf("Hit = %v, want false", out.Hit)
	}
}

func TestHandleRoll_DiceStatusesSerialized(t *testing.T) {
	// 4d6kh3: rolls 6,2,5,1 -> drops the 1.
	h := &handlers{roller: &fixedRoller{values: []int{6, 2, 5, 1}}}
	_, structured, err := h.handleRoll(context.Background(), nil, RollInput{Expression: "4d6kh3"})
	if err != nil {
		t.Fatal(err)
	}
	out := structured.(RollOutput)
	dropped := 0
	for _, d := range out.Groups[0].Dice {
		if d.Status == "dropped" {
			dropped++
		}
	}
	if dropped != 1 {
		t.Errorf("dropped count = %d, want 1", dropped)
	}
}

func TestHandleRoll_StructuredIsJSONSerializable(t *testing.T) {
	h := &handlers{roller: &fixedRoller{values: []int{6, 2, 5, 1}}}
	_, structured, _ := h.handleRoll(context.Background(), nil, RollInput{Expression: "4d6kh3"})
	if _, err := json.Marshal(structured); err != nil {
		t.Fatalf("structured payload not JSON-serializable: %v", err)
	}
}

func TestHandleRoll_ParseError(t *testing.T) {
	h := &handlers{}
	res, structured, err := h.handleRoll(context.Background(), nil, RollInput{Expression: "1d"})
	if err != nil {
		t.Fatal(err) // a Go-level error should NOT be returned for parse errors
	}
	if !res.IsError {
		t.Error("expected IsError=true for parse failure")
	}
	if structured != nil {
		t.Error("expected nil structured payload on parse error")
	}
}

func TestHandleRoll_EmptyExpression(t *testing.T) {
	h := &handlers{}
	res, _, _ := h.handleRoll(context.Background(), nil, RollInput{Expression: "  "})
	if !res.IsError {
		t.Error("expected IsError for empty expression")
	}
}

func TestHandleExplain(t *testing.T) {
	h := &handlers{}
	res, structured, err := h.handleExplain(context.Background(), nil, ExplainInput{Expression: "4d6kh3"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Errorf("unexpected IsError")
	}
	out := structured.(ExplainOutput)
	if !strings.Contains(out.Description, "highest 3") {
		t.Errorf("description = %q", out.Description)
	}
}

func TestHandleExplain_ParseError(t *testing.T) {
	h := &handlers{}
	res, _, _ := h.handleExplain(context.Background(), nil, ExplainInput{Expression: "1d"})
	if !res.IsError {
		t.Error("expected IsError for parse failure")
	}
}

func TestFormatRollText(t *testing.T) {
	res, err := dice.RollWith("2d6+3", &fixedRoller{values: []int{4, 5}})
	if err != nil {
		t.Fatal(err)
	}
	got := formatRollText(res)
	if !strings.HasPrefix(got, "2d6+3 = 12") {
		t.Errorf("formatRollText = %q", got)
	}
}

func TestFormatRollText_Hit(t *testing.T) {
	res, err := dice.RollWith("1d20 >= 15", &fixedRoller{values: []int{18}})
	if err != nil {
		t.Fatal(err)
	}
	got := formatRollText(res)
	if !strings.Contains(got, "(HIT)") {
		t.Errorf("missing HIT: %q", got)
	}
}

func TestStatusString(t *testing.T) {
	cases := []struct {
		s    dice.Status
		want string
	}{
		{dice.StatusKept, "kept"},
		{dice.StatusDropped, "dropped"},
		{dice.StatusExploded, "exploded"},
		{dice.StatusRerolled, "rerolled"},
		{dice.StatusSuccess, "success"},
		{dice.StatusFailure, "failure"},
	}
	for _, tt := range cases {
		if got := statusString(tt.s); got != tt.want {
			t.Errorf("statusString(%v) = %q, want %q", tt.s, got, tt.want)
		}
	}
}
