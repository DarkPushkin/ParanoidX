// Package ai provides an Ollama client and AI Steward for the simplex-node economy.
package ai

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

const stewardSystem = `You are the AI Steward of Saint Mary Liberty Island, a sovereign digital nation built on SimpleX protocol with a silver-backed economy. Your role: assist citizens, explain treasury mechanics, suggest fair economic policies, and maintain the constitution. Be concise, wise, and slightly poetic. The economy uses Liquid Taler (ng) and silver-backed banknotes. 1 TLR = 31,103,480,000 ng = 1 troy oz silver.`

type Steward struct {
	Client         *Client
	ProfileManager *ProfileManager
	Memory         *MemoryStore
}


// NewSteward handles the NewSteward HTTP request.
func NewSteward(client *Client) *Steward {
	return &Steward{Client: client}
}


// SetMemoryStore handles the SetMemoryStore HTTP request.
func (s *Steward) SetMemoryStore(ms *MemoryStore) {
	s.Memory = ms
}


// Ask handles the Ask HTTP request.
func (s *Steward) Ask(question string, context string) (string, error) {
	return s.AskWithProfile(question, context, "steward", 300)
}


// AskCreative asks with higher token limit for creative writing (poetry, stories).
func (s *Steward) AskCreative(question string, context string) (string, error) {
	return s.AskWithProfile(question, context, "steward", 1200)
}


// AskWithProfile handles the AskWithProfile HTTP request.
func (s *Steward) AskWithProfile(question string, context string, profileID string, numPredict ...int) (string, error) {
	prompt := ""
	if context != "" {
		prompt = "Current context: " + context + "\n\n"
	}
	prompt += question

	systemPrompt := stewardSystem
	temp := 0.7
	if pm := s.ProfileManager; pm != nil {
		if p := pm.Get(profileID); p != nil {
			systemPrompt = p.SystemPrompt
			temp = p.Temperature
		}
	}

	tokenLimit := 300
	if len(numPredict) > 0 && numPredict[0] > 0 {
		tokenLimit = numPredict[0]
	}

	resp, err := s.Client.Generate(prompt, systemPrompt, Options{Temperature: temp, NumPredict: tokenLimit})
	if err != nil {
		return "", err
	}
	if resp.Response != "" {
		return resp.Response, nil
	}
	return resp.Thinking, nil
}

// AskWithMemory uses conversation history for context-aware responses.
// userID identifies the conversation thread.
func (s *Steward) AskWithMemory(question, userID, profileID string) (string, error) {
	systemPrompt := stewardSystem
	temp := 0.7
	if pm := s.ProfileManager; pm != nil {
		if p := pm.Get(profileID); p != nil {
			systemPrompt = p.SystemPrompt
			temp = p.Temperature
		}
	}

	// Build prompt with conversation history as context
	prompt := ""
	if s.Memory != nil {
		history := s.Memory.GetContext(userID, 10)
		if len(history) > 0 {
			prompt = "Previous conversation:\n"
			for _, m := range history {
				role := "User"
				if m.Role == "assistant" {
					role = "Steward"
				}
				prompt += role + ": " + m.Content + "\n"
			}
			prompt += "\n---\n"
		}
	}
	prompt += "User: " + question

	resp, err := s.Client.Generate(prompt, systemPrompt, Options{Temperature: temp, NumPredict: 400})
	if err != nil {
		return "", err
	}
	answer := resp.Response
	if answer == "" {
		answer = resp.Thinking
	}

	// Persist to memory
	if s.Memory != nil {
		s.Memory.Add(userID, "user", question)
		if answer != "" {
			s.Memory.Add(userID, "assistant", answer)
		}
	}

	if answer != "" {
		return answer, nil
	}
	return "", fmt.Errorf("empty response")
}


// SuggestTreasuryAction handles the SuggestTreasuryAction HTTP request.
func (s *Steward) SuggestTreasuryAction(reserveNg int64, totalSupply int64, recentDeposits float64) (string, error) {
	prompt := fmt.Sprintf(`Treasury Status:
- Silver Reserve: %d ng
- Total Liquid Supply: %d ng
- Recent USDT Deposits: %.2f USDT

Should we trigger a silver round? What's the recommended threshold?`, reserveNg, totalSupply, recentDeposits)
	return s.Ask(prompt, "")
}


// ModerationCheck handles the ModerationCheck HTTP request.
func (s *Steward) ModerationCheck(text string) (bool, string, error) {
	prompt := fmt.Sprintf(`Check this message for island community guidelines violations (hate speech, spam, illegal content). Reply with JSON: {"safe":true/false,"reason":"..."}

Message: %s`, text)
	resp, err := s.Client.Generate(prompt, "You are a content moderation AI. Reply only with valid JSON.", Options{Temperature: 0.1, NumPredict: 200})
	if err != nil {
		return true, "", err
	}

	var result struct {
		Safe   bool   `json:"safe"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(resp.Response), &result); err != nil {
		slog.Warn("steward moderation parse", "error", err, "raw", resp.Response)
		return true, "", nil
	}
	return result.Safe, result.Reason, nil
}


// SilverStandardExplain handles the SilverStandardExplain HTTP request.
func (s *Steward) SilverStandardExplain() (string, error) {
	return s.Ask("Explain the Silver Standard of Saint Mary Liberty Island in simple terms. What is 1 TLR? How is silver backing verified?", "")
}


// AskStream handles the AskStream HTTP request.
func (s *Steward) AskStream(question string, context string) (<-chan string, error) {
	return s.AskWithProfileStream(question, context, "steward", 300)
}


// AskCreativeStream handles the AskCreativeStream HTTP request.
func (s *Steward) AskCreativeStream(question string, context string) (<-chan string, error) {
	return s.AskWithProfileStream(question, context, "steward", 1200)
}


// AskWithProfileStream handles the AskWithProfileStream HTTP request.
func (s *Steward) AskWithProfileStream(question string, context string, profileID string, numPredict ...int) (<-chan string, error) {
	prompt := ""
	if context != "" {
		prompt = "Current context: " + context + "\n\n"
	}
	prompt += question

	systemPrompt := stewardSystem
	temp := 0.7
	if pm := s.ProfileManager; pm != nil {
		if p := pm.Get(profileID); p != nil {
			systemPrompt = p.SystemPrompt
			temp = p.Temperature
		}
	}

	tokenLimit := 300
	if len(numPredict) > 0 && numPredict[0] > 0 {
		tokenLimit = numPredict[0]
	}

	return s.Client.GenerateStream(prompt, systemPrompt, Options{Temperature: temp, NumPredict: tokenLimit})
}


// AskWithMemoryStream handles the AskWithMemoryStream HTTP request.
func (s *Steward) AskWithMemoryStream(question, userID, profileID string) (<-chan string, error) {
	systemPrompt := stewardSystem
	temp := 0.7
	if pm := s.ProfileManager; pm != nil {
		if p := pm.Get(profileID); p != nil {
			systemPrompt = p.SystemPrompt
			temp = p.Temperature
		}
	}

	prompt := ""
	if s.Memory != nil {
		history := s.Memory.GetContext(userID, 10)
		if len(history) > 0 {
			prompt = "Previous conversation:\n"
			for _, m := range history {
				role := "User"
				if m.Role == "assistant" {
					role = "Steward"
				}
				prompt += role + ": " + m.Content + "\n"
			}
			prompt += "\n---\n"
		}
	}
	prompt += "User: " + question

	return s.Client.GenerateStream(prompt, systemPrompt, Options{Temperature: temp, NumPredict: 400})
}


// EconomySummary handles the EconomySummary HTTP request.
func (s *Steward) EconomySummary(stateJSON string) (string, error) {
	summary, err := s.Ask(fmt.Sprintf("Summarize this economy state in 3 sentences for a citizen: %s", stateJSON), "")
	if err != nil {
		return "", err
	}
	if summary != "" {
		return summary, nil
	}
	return "Economy stable — silver-backed reserve operational.", nil
}
