package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/oldschoolgamers/dice"
)

// fixedRoller mirrors the helper in the library tests; we avoid importing
// dice_test because it's a black-box test package.
type fixedRoller struct {
	values []int
	i      int
}

func (r *fixedRoller) RollDie(int) (int, error) {
	v := r.values[r.i]
	r.i++
	return v, nil
}

func render(t *testing.T, expr string, rolls []int, color bool) string {
	t.Helper()
	res, err := dice.RollWith(expr, &fixedRoller{values: rolls})
	if err != nil {
		t.Fatalf("RollWith(%q): %v", expr, err)
	}
	var buf bytes.Buffer
	Render(&buf, color, res)
	return buf.String()
}

func TestRender_PlainBasic(t *testing.T) {
	got := render(t, "2d6+3", []int{4, 5}, false)
	want := "2d6+3 = 12\n  2d6: 4 5\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRender_PlainKeepDrop(t *testing.T) {
	got := render(t, "4d6kh3", []int{6, 2, 5, 1}, false)
	if !strings.Contains(got, "= 13") {
		t.Errorf("missing total 13: %s", got)
	}
	if !strings.Contains(got, "[1]") {
		t.Errorf("dropped die not bracketed: %s", got)
	}
}

func TestRender_PlainAdvantage(t *testing.T) {
	got := render(t, "adv(1d20)", []int{9, 15}, false)
	// Winner first (15), then loser (9) marked dropped.
	if !strings.Contains(got, "= 15") {
		t.Errorf("missing winner total: %s", got)
	}
	if !strings.Contains(got, "[9]") {
		t.Errorf("loser not dimmed/bracketed: %s", got)
	}
}

func TestRender_PlainExplodeMarker(t *testing.T) {
	got := render(t, "1d6x", []int{6, 4}, false)
	if !strings.Contains(got, "6!") {
		t.Errorf("exploded die missing '!' marker: %s", got)
	}
}

func TestRender_PlainSuccessFailure(t *testing.T) {
	got := render(t, "5d10cs>=8cf=1", []int{8, 1, 9, 1, 5}, false)
	if !strings.Contains(got, "8*") || !strings.Contains(got, "9*") {
		t.Errorf("missing success markers: %s", got)
	}
	if !strings.Contains(got, "1x") {
		t.Errorf("missing failure marker: %s", got)
	}
}

func TestRender_PlainHitMiss(t *testing.T) {
	hit := render(t, "1d20 >= 15", []int{18}, false)
	if !strings.Contains(hit, "HIT") {
		t.Errorf("missing HIT label: %s", hit)
	}
	miss := render(t, "1d20 >= 15", []int{8}, false)
	if !strings.Contains(miss, "MISS") {
		t.Errorf("missing MISS label: %s", miss)
	}
}

func TestRender_ColorOnHit(t *testing.T) {
	got := render(t, "1d20 >= 15", []int{18}, true)
	if !strings.Contains(got, "\x1b[32m") {
		t.Errorf("expected green ANSI on hit: %q", got)
	}
}

func TestRender_ColorOnMiss(t *testing.T) {
	got := render(t, "1d20 >= 15", []int{8}, true)
	if !strings.Contains(got, "\x1b[31m") {
		t.Errorf("expected red ANSI on miss: %q", got)
	}
}

func TestRender_NoColorWhenDisabled(t *testing.T) {
	got := render(t, "1d20 >= 15", []int{18}, false)
	if strings.Contains(got, "\x1b[") {
		t.Errorf("plain output contained ANSI: %q", got)
	}
}
