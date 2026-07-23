// Package steward implements the Steward AI service
package steward

import "log/slog"

// Deviation describes a detected rule deviation.
type Deviation struct {
	Rule      string  `json:"rule"`
	Value     float64 `json:"value"`
	Target    float64 `json:"target"`
	Severity  string  `json:"severity"` // minor / major / critical
	Message   string  `json:"message"`
}

// Analyze checks all metrics against the constitution.
func Analyze(constitution *Constitution, metrics *StewardMetrics) []Deviation {
	if metrics == nil {
		return nil
	}

	var deviations []Deviation

	deviations = append(deviations, checkRatio(constitution, "silver_reserve_ratio", metrics.ReserveRatio)...)
	deviations = append(deviations, checkRatio(constitution, "treasury_tier_fat_threshold", float64(metrics.MonthlyOpsNg))...)

	for _, rule := range constitution.Rules {
		switch rule.Name {
		case "silver_reserve_ratio":
		case "treasury_tier_fat_threshold":
		case "treasury_tier_veryfat_threshold":
		case "ng_per_tlr":
		default:
		}
	}

	return deviations
}

func checkRatio(constitution *Constitution, name string, value float64) []Deviation {
	dev, sev := constitution.CheckRule(name, value)
	if sev == "unknown" || (sev == "minor" && abs(dev) < 0.1) {
		return nil
	}

	msg := name + " is "
	if dev > 0 {
		msg += "above target"
	} else {
		msg += "below target"
	}

	slog.Warn("steward deviation detected", "rule", name, "value", value, "severity", sev, "deviation", dev)

	return []Deviation{{
		Rule:     name,
		Value:    value,
		Target:   0,
		Severity: sev,
		Message:  msg,
	}}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// Decision is an action the steward recommends.
type Decision struct {
	Rule      string `json:"rule"`
	Action    string `json:"action"`    // e.g. "adjust_commission", "notify_admin", "lock_market"
	Parameter string `json:"parameter"` // which parameter to change
	NewValue  string `json:"new_value"`
	Reason    string `json:"reason"`
}

// Decide generates decisions from deviations.
func Decide(constitution *Constitution, deviations []Deviation) []Decision {
	var decisions []Decision

	for _, d := range deviations {
		switch d.Severity {
		case "minor":
			decisions = append(decisions, Decision{
				Rule:   d.Rule,
				Action: "auto_adjust",
				Reason: d.Message + " — auto-adjusted within safe bounds",
			})
		case "major":
			decisions = append(decisions, Decision{
				Rule:   d.Rule,
				Action: "notify_admin",
				Reason: d.Message + " — requires admin attention",
			})
		case "critical":
			decisions = append(decisions, Decision{
				Rule:   d.Rule,
				Action: "require_consensus",
				Reason: d.Message + " — CRITICAL: requires multi-sig consensus",
			})
		}

		// Auto-recovery: if reserve ratio drops critically, restrict minting
		if d.Rule == "silver_reserve_ratio" && d.Severity == "critical" && d.Value < 0.60 {
			decisions = append(decisions, Decision{
				Rule:   "auto_recovery",
				Action: "auto_adjust",
				Reason: "Reserve ratio critically low — auto-restricting minting",
			})
		}
	}

	return decisions
}
