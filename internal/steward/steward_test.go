// Package steward implements the Steward AI service
package steward

import (
	"testing"
)


// TestDefaultConstitutionHasRules handles the TestDefaultConstitutionHasRules HTTP request.
func TestDefaultConstitutionHasRules(t *testing.T) {
	c := DefaultConstitution()
	if len(c.Rules) == 0 {
		t.Fatal("expected rules")
	}
}


// TestCheckRuleOnTarget handles the TestCheckRuleOnTarget HTTP request.
func TestCheckRuleOnTarget(t *testing.T) {
	c := DefaultConstitution()
	dev, sev := c.CheckRule("silver_reserve_ratio", 0.70)
	if sev != "minor" {
		t.Errorf("expected minor, got %s", sev)
	}
	if abs(dev) > 0.05 {
		t.Errorf("expected near-zero deviation, got %f", dev)
	}
}


// TestCheckRuleCritical handles the TestCheckRuleCritical HTTP request.
func TestCheckRuleCritical(t *testing.T) {
	c := DefaultConstitution()
	_, sev := c.CheckRule("silver_reserve_ratio", 0.50)
	if sev != "critical" {
		t.Errorf("expected critical for 0.50 ratio, got %s", sev)
	}
}


// TestCheckRuleUnknown handles the TestCheckRuleUnknown HTTP request.
func TestCheckRuleUnknown(t *testing.T) {
	c := DefaultConstitution()
	_, sev := c.CheckRule("nonexistent", 42)
	if sev != "unknown" {
		t.Errorf("expected unknown, got %s", sev)
	}
}


// TestAnalyzeNoMetrics handles the TestAnalyzeNoMetrics HTTP request.
func TestAnalyzeNoMetrics(t *testing.T) {
	deviations := Analyze(DefaultConstitution(), nil)
	if deviations != nil {
		t.Errorf("expected nil for nil metrics")
	}
}


// TestDecideMinor handles the TestDecideMinor HTTP request.
func TestDecideMinor(t *testing.T) {
	c := DefaultConstitution()
	deviations := []Deviation{
		{Rule: "test", Severity: "minor", Message: "slightly off"},
	}
	decisions := Decide(c, deviations)
	if len(decisions) != 1 {
		t.Fatal("expected 1 decision")
	}
	if decisions[0].Action != "auto_adjust" {
		t.Errorf("expected auto_adjust, got %s", decisions[0].Action)
	}
}


// TestDecideMajor handles the TestDecideMajor HTTP request.
func TestDecideMajor(t *testing.T) {
	c := DefaultConstitution()
	deviations := []Deviation{
		{Rule: "test", Severity: "major", Message: "significantly off"},
	}
	decisions := Decide(c, deviations)
	if len(decisions) != 1 {
		t.Fatal("expected 1 decision")
	}
	if decisions[0].Action != "notify_admin" {
		t.Errorf("expected notify_admin, got %s", decisions[0].Action)
	}
}


// TestDecideCritical handles the TestDecideCritical HTTP request.
func TestDecideCritical(t *testing.T) {
	c := DefaultConstitution()
	deviations := []Deviation{
		{Rule: "test", Severity: "critical", Message: "way off"},
	}
	decisions := Decide(c, deviations)
	if decisions[0].Action != "require_consensus" {
		t.Errorf("expected require_consensus, got %s", decisions[0].Action)
	}
}


// TestConstitutionRuleBounds handles the TestConstitutionRuleBounds HTTP request.
func TestConstitutionRuleBounds(t *testing.T) {
	c := DefaultConstitution()
	for _, rule := range c.Rules {
		if rule.Min >= rule.Max {
			t.Errorf("rule %s: min %f >= max %f", rule.Name, rule.Min, rule.Max)
		}
		if rule.Target < rule.Min || rule.Target > rule.Max {
			t.Errorf("rule %s: target %f outside [%f, %f]", rule.Name, rule.Target, rule.Min, rule.Max)
		}
	}
}
