package audit

import "testing"

// Policy() states, as data, what textLevel and the add() calls in Flavor say as
// code. Publishing it is what lets the website print the floors instead of
// re-typing them — and a mirrored policy is a policy that can be wrong in public.
//
// So it is checked against the measurements themselves: every pair the audit
// takes has to be claimed by exactly one published rule, at the same floor and
// the same severity. A floor changed in one place and not the other fails here.
func TestThePublishedPolicyIsThePolicyApplied(t *testing.T) {
	pal := mustLoad(t)
	rep := Flavor(pal, pal.Flavors[0])

	for _, c := range rep.Checks {
		rule, ok := RuleFor(c.FG, c.BG)
		if !ok {
			t.Errorf("%s on %s is measured, but no published rule claims it", c.FG, c.BG)
			continue
		}
		if c.Want != rule.Floor {
			t.Errorf("%s on %s was measured against %.1f; the published rule for %q says %.1f",
				c.FG, c.BG, c.Want, rule.Subject, rule.Floor)
		}
		if c.Fatal != rule.Gate {
			t.Errorf("%s on %s has fatal=%v; the published rule for %q says gate=%v",
				c.FG, c.BG, c.Fatal, rule.Subject, rule.Gate)
		}
	}
}

// The other direction: a rule nothing measures is a row the website would print
// about a check that does not happen.
func TestEveryPublishedRuleIsExercised(t *testing.T) {
	pal := mustLoad(t)
	rep := Flavor(pal, pal.Flavors[0])

	for i, rule := range Policy() {
		used := false
		for _, c := range rep.Checks {
			if got, ok := RuleFor(c.FG, c.BG); ok && got.Subject == rule.Subject && got.Floor == rule.Floor {
				used = true
				break
			}
		}
		if !used {
			t.Errorf("rule %d (%s, floor %.1f) is published but never applied", i, rule.Subject, rule.Floor)
		}
	}
}

// The published count has to be what every flavour really measures, not just the
// first — the summary reports one number for all of them.
func TestTheSummaryCountsWhatEveryFlavourMeasures(t *testing.T) {
	pal := mustLoad(t)
	s := Summarise(pal)

	for _, f := range pal.Flavors {
		if got := len(Flavor(pal, f).Checks); got != s.PairsPerFlavor {
			t.Errorf("flavour %q measures %d pairs, the summary publishes %d", f.ID, got, s.PairsPerFlavor)
		}
	}
	if len(s.Flavors) != len(pal.Flavors) {
		t.Errorf("summary lists %d flavours, the palette has %d", len(s.Flavors), len(pal.Flavors))
	}
	if len(s.Policy) == 0 {
		t.Error("the summary publishes no policy")
	}
}
