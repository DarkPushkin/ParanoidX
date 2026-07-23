// Package ai provides AI integration with Ollama, including chat, generation, and monitoring
package ai

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// StewardMonitor periodically checks system health against the constitution.
type StewardMonitor struct {
	mu        sync.Mutex
	Client    *Client
	Decisions []MonitorDecision
}

type MonitorDecision struct {
	Check     string `json:"check"`
	Status    string `json:"status"` // ok, warning, critical
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}


// NewStewardMonitor handles the NewStewardMonitor HTTP request.
func NewStewardMonitor(client *Client) *StewardMonitor {
	return &StewardMonitor{Client: client}
}

// CheckReserveRatio verifies the silver reserve covers all banknote liabilities.
func (sm *StewardMonitor) CheckReserveRatio(reserveNg, liabilitiesNg int64) MonitorDecision {
	var ratio float64
	status := "ok"
	if liabilitiesNg > 0 {
		ratio = float64(reserveNg) / float64(liabilitiesNg)
		if ratio < 1.0 {
			status = "critical"
		} else if ratio < 1.5 {
			status = "warning"
		}
	}
	msg := fmt.Sprintf("Reserve ratio: %.4f (reserve=%d ng, liabilities=%d ng)", ratio, reserveNg, liabilitiesNg)
	return MonitorDecision{Check: "reserve_ratio", Status: status, Message: msg}
}

// CheckTreasuryTier monitors the treasury tier for deflation triggers.
func (sm *StewardMonitor) CheckTreasuryTier(tierName string, burnThresholdExceeded bool) MonitorDecision {
	msg := fmt.Sprintf("Treasury tier: %s", tierName)
	status := "ok"
	if tierName == "very_fat" && burnThresholdExceeded {
		status = "action_required"
		msg += " — deflation burn should be triggered"
	}
	return MonitorDecision{Check: "treasury_tier", Status: status, Message: msg}
}

// CheckActiveServices monitors registered service availability.
func (sm *StewardMonitor) CheckActiveServices(activeCount, totalCount int) MonitorDecision {
	msg := fmt.Sprintf("Services: %d/%d active", activeCount, totalCount)
	status := "ok"
	if totalCount > 0 && float64(activeCount)/float64(totalCount) < 0.5 {
		status = "critical"
		msg += " — less than 50% of services active"
	} else if activeCount < totalCount {
		status = "warning"
		msg += " — some services are offline"
	}
	return MonitorDecision{Check: "services", Status: status, Message: msg}
}

// Record saves a monitor decision.
func (sm *StewardMonitor) Record(d MonitorDecision) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.Decisions = append(sm.Decisions, d)
	slog.Info("steward monitor", "check", d.Check, "status", d.Status, "message", d.Message)
}

// Summary returns recent monitor decisions.
func (sm *StewardMonitor) Summary(n int) []MonitorDecision {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if n <= 0 || n > len(sm.Decisions) {
		n = len(sm.Decisions)
	}
	out := make([]MonitorDecision, n)
	copy(out, sm.Decisions[len(sm.Decisions)-n:])
	return out
}

// AskWithConstitution adds constitution context to the steward's knowledge.
func (s *Steward) AskWithConstitution(question string, constitutionText string, stateJSON string) (string, error) {
	messages := []Message{
		{Role: "system", Content: stewardSystem},
	}

	if constitutionText != "" {
		articles := strings.Split(constitutionText, "\n")
		text := "Island Constitution:\n"
		for _, a := range articles {
			text += a + "\n"
		}
		messages = append(messages, Message{Role: "system", Content: text})
	}

	if stateJSON != "" {
		messages = append(messages, Message{Role: "system", Content: "Current system state: " + stateJSON})
	}

	messages = append(messages, Message{Role: "user", Content: question})

	resp, err := s.Client.Chat(messages, Options{Temperature: 0.5, NumPredict: 600})
	if err != nil {
		return "", err
	}
	return resp.Message.Content, nil
}
