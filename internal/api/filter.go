// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var GlobalContentFilter *ContentFilterEngine

// ContentFilterRule defines a single content filtering rule with word, action, and optional replacement.
type ContentFilterRule struct {
	Word        string `json:"word"`
	Action      string `json:"action"`      // block, replace, flag
	Replacement string `json:"replacement"` // used if action=replace
}

// ContentFilterEngine manages a set of content filtering rules with thread-safe access.
type ContentFilterEngine struct {
	mu    sync.RWMutex
	path  string
	rules []ContentFilterRule
}


// NewContentFilterEngine creates a content filter engine and loads rules from disk with defaults.
func NewContentFilterEngine(dataDir string) *ContentFilterEngine {
	fe := &ContentFilterEngine{
		path: filepath.Join(dataDir, "content_filter_rules.json"),
	}
	data, err := os.ReadFile(fe.path)
	if err == nil {
		json.Unmarshal(data, &fe.rules)
	}
	// Default rules
	if len(fe.rules) == 0 {
		fe.rules = append(fe.rules,
			ContentFilterRule{Word: "fuck", Action: "block"},
			ContentFilterRule{Word: "shit", Action: "block"},
			ContentFilterRule{Word: "asshole", Action: "block"},
			ContentFilterRule{Word: "porn", Action: "flag"},
			ContentFilterRule{Word: "spam", Action: "flag"},
			ContentFilterRule{Word: "scam", Action: "block"},
		)
		fe.save()
	}
	return fe
}

func (fe *ContentFilterEngine) save() {
	data, _ := json.MarshalIndent(fe.rules, "", "  ")
	os.WriteFile(fe.path, data, 0644)
}


// AddRule adds or updates a content filtering rule.
func (fe *ContentFilterEngine) AddRule(word, action, replacement string) {
	fe.mu.Lock()
	defer fe.mu.Unlock()
	for i, r := range fe.rules {
		if r.Word == word {
			fe.rules[i].Action = action
			fe.rules[i].Replacement = replacement
			fe.save()
			return
		}
	}
	fe.rules = append(fe.rules, ContentFilterRule{Word: word, Action: action, Replacement: replacement})
	fe.save()
}


// RemoveRule deletes a content filtering rule by word.
func (fe *ContentFilterEngine) RemoveRule(word string) {
	fe.mu.Lock()
	defer fe.mu.Unlock()
	n := 0
	for _, r := range fe.rules {
		if r.Word != word {
			fe.rules[n] = r
			n++
		}
	}
	fe.rules = fe.rules[:n]
	fe.save()
}


// GetRules returns a copy of all content filtering rules.
func (fe *ContentFilterEngine) GetRules() []ContentFilterRule {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	out := make([]ContentFilterRule, len(fe.rules))
	copy(out, fe.rules)
	return out
}

// Filter returns (filteredText, note, blocked)
func (fe *ContentFilterEngine) Filter(text string) (string, string, bool) {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	lower := strings.ToLower(text)
	for _, rule := range fe.rules {
		if strings.Contains(lower, rule.Word) {
			switch rule.Action {
			case "block":
				return "", "blocked: " + rule.Word, true
			case "replace":
				if rule.Replacement != "" {
					re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(rule.Word))
					text = re.ReplaceAllString(text, rule.Replacement)
				}
			case "flag":
				return text, "flagged: " + rule.Word, true
			}
		}
	}
	return text, "", false
}


// ContentFilterRulesHandler manages content filter rules via POST (add/remove) and GET (list).
func ContentFilterRulesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				Word        string `json:"word"`
				Action      string `json:"action"`
				Replacement string `json:"replacement"`
				Remove      string `json:"remove"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad json", 400)
				return
			}
			if req.Remove != "" {
				GlobalContentFilter.RemoveRule(req.Remove)
			} else if req.Word != "" {
				GlobalContentFilter.AddRule(req.Word, req.Action, req.Replacement)
			}
		}
		writeJSON(w, map[string]any{"ok": true, "rules": GlobalContentFilter.GetRules()})
	}
}


// ContentFilterTestHandler tests a text string against the content filter rules.
func ContentFilterTestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		text := r.URL.Query().Get("text")
		if text == "" {
			http.Error(w, "text required", 400)
			return
		}
		filtered, note, blocked := GlobalContentFilter.Filter(text)
		writeJSON(w, map[string]any{
			"ok":       true,
			"blocked":  blocked,
			"filtered": filtered,
			"note":     note,
		})
	}
}


// ContentFilterCheck returns whether the given text is blocked by the filter.
func ContentFilterCheck(text string) (bool, string) {
	if GlobalContentFilter == nil {
		return false, ""
	}
	_, note, blocked := GlobalContentFilter.Filter(text)
	return blocked, note
}


// ContentFilterReload adds multiple words as blocked rules to the content filter.
func ContentFilterReload(words []string) {
	if GlobalContentFilter == nil {
		return
	}
	for _, w := range words {
		GlobalContentFilter.AddRule(w, "block", "")
	}
}
