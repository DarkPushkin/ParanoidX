// Package economy implements the island economy system
package economy

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// WheelRewardTier defines the rarity of a wheel spin result.
type WheelRewardTier string

const (
	WheelCommon    WheelRewardTier = "common"
	WheelRare      WheelRewardTier = "rare"
	WheelEpic      WheelRewardTier = "epic"
	WheelLegendary WheelRewardTier = "legendary"
)

// WheelReward is the result of a spin.
type WheelReward struct {
	Tier    WheelRewardTier `json:"tier"`
	Label   string          `json:"label"`
	NgAward int64           `json:"ng_award"`
}

var wheelWeights = map[WheelRewardTier]int{
	WheelCommon:    600, // 60%
	WheelRare:      250, // 25%
	WheelEpic:      100, // 10%
	WheelLegendary: 50,  // 5%
}

// wheelPrizes defines possible prizes per tier.
var wheelPrizes = map[WheelRewardTier][]struct {
	label string
	ng    int64
}{
	WheelCommon: {
		{"1 TLR", 1 * NGPerTLR},
		{"2 TLR", 2 * NGPerTLR},
		{"5 TLR", 5 * NGPerTLR},
		{"1 day dividend boost", 0},
	},
	WheelRare: {
		{"10 TLR", 10 * NGPerTLR},
		{"25 TLR", 25 * NGPerTLR},
		{"3 day dividend boost", 0},
	},
	WheelEpic: {
		{"50 TLR", 50 * NGPerTLR},
		{"100 TLR", 100 * NGPerTLR},
		{"Common banknote pack", 0},
	},
	WheelLegendary: {
		{"250 TLR", 250 * NGPerTLR},
		{"500 TLR", 500 * NGPerTLR},
		{"Rare banknote", 0},
	},
}

// pickTier rolls a weighted random tier.
func pickTier() WheelRewardTier {
	total := 0
	for _, w := range wheelWeights {
		total += w
	}
	r := rand.Intn(total)
	cum := 0
	for tier, w := range wheelWeights {
		cum += w
		if r < cum {
			return tier
		}
	}
	return WheelCommon
}

// SpinWheel runs a single spin and returns the reward.
func SpinWheel() WheelReward {
	tier := pickTier()
	prizes := wheelPrizes[tier]
	pick := prizes[rand.Intn(len(prizes))]
	return WheelReward{Tier: tier, Label: pick.label, NgAward: pick.ng}
}

// WheelSpinner manages per-pubkey daily spins.
type WheelSpinner struct {
	mu       sync.Mutex
	lastSpin map[string]time.Time // pubkey → last spin time
}


// NewWheelSpinner handles the NewWheelSpinner HTTP request.
func NewWheelSpinner() *WheelSpinner {
	return &WheelSpinner{lastSpin: make(map[string]time.Time)}
}

// CanSpin returns true if the pubkey hasn't spun today.
func (ws *WheelSpinner) CanSpin(pubkey string) bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	last, ok := ws.lastSpin[pubkey]
	if !ok {
		return true
	}
	now := time.Now()
	// Allow if last spin was on a different day
	return last.Year() != now.Year() || last.YearDay() != now.YearDay()
}

// Spin attempts a daily spin. Returns error if already spun today.
func (ws *WheelSpinner) Spin(pubkey string) (WheelReward, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	now := time.Now()
	last, ok := ws.lastSpin[pubkey]
	if ok && last.Year() == now.Year() && last.YearDay() == now.YearDay() {
		return WheelReward{}, fmt.Errorf("already spun today")
	}

	reward := SpinWheel()
	ws.lastSpin[pubkey] = now
	return reward, nil
}

// NextSpinTime returns when the pubkey can spin again.
func (ws *WheelSpinner) NextSpinTime(pubkey string) time.Time {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	last, ok := ws.lastSpin[pubkey]
	if !ok {
		return time.Time{}
	}
	next := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, last.Location()).Add(24 * time.Hour)
	return next
}
