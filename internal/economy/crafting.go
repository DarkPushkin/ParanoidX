// Package economy implements the island economy system
package economy

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// CraftingManager handles 5→1 banknote upgrades.
type CraftingManager struct {
	mu sync.Mutex
}

// NewCraftingManager creates an empty crafting manager.
func NewCraftingManager() *CraftingManager {
	return &CraftingManager{}
}

// NextRarity returns the rarity one step above the given one.
func NextRarity(current string) (string, error) {
	order := []string{"common", "rare", "epic", "legendary", "genesis"}
	for i, r := range order {
		if r == current && i < len(order)-1 {
			return order[i+1], nil
		}
	}
	return "", fmt.Errorf("cannot upgrade %q: already max rarity or unknown", current)
}

// UpgradeInputCount is the number of banknotes needed for one upgrade (5).
const UpgradeInputCount = 5

// CraftUpgrade consumes 5 banknotes of the same rarity and creates 1 of the next tier.
func (cm *CraftingManager) CraftUpgrade(dataDir, holder string, serials []string) ([]string, *BanknoteV2, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(serials) != UpgradeInputCount {
		return nil, nil, fmt.Errorf("need exactly %d banknotes, got %d", UpgradeInputCount, len(serials))
	}

	banknotes, err := LoadBanknotesV2(dataDir)
	if err != nil {
		return nil, nil, fmt.Errorf("load banknotes: %w", err)
	}

	var toBurn []BanknoteV2
	seen := map[string]bool{}
	var rarity string

	for _, s := range serials {
		if seen[s] {
			return nil, nil, fmt.Errorf("duplicate serial: %s", s)
		}
		seen[s] = true
		found := false
		for _, b := range banknotes {
			if b.Serial == s && b.Holder == holder && b.Status == "active" {
				if rarity == "" {
					rarity = b.Rarity
				} else if b.Rarity != rarity {
					return nil, nil, fmt.Errorf("mixed rarities: %s and %s", rarity, b.Rarity)
				}
				toBurn = append(toBurn, b)
				found = true
				break
			}
		}
		if !found {
			return nil, nil, fmt.Errorf("banknote %s not found or not owned/active", s)
		}
	}

	nextRarity, err := NextRarity(rarity)
	if err != nil {
		return nil, nil, err
	}

	now := fmt.Sprintf("%d", time.Now().Unix())
	upgraded := BanknoteV2{
		Serial:         fmt.Sprintf("MB-%s-CRAFT-%s", strings.ToUpper(nextRarity), now),
		DenominationNg: sumDenominations(toBurn) * 2,
		Rarity:         nextRarity,
		Multiplier:     rarityMultiplier[nextRarity],
		Holder:         holder,
		FrozenNg:       sumDenominations(toBurn) * 2,
		Status:         "active",
		MintedAt:       now,
	}

	for i := range banknotes {
		for _, s := range serials {
			if banknotes[i].Serial == s {
				banknotes[i].Status = "burned"
			}
		}
	}

	banknotes = append(banknotes, upgraded)
	SaveBanknotesV2(dataDir, banknotes)
	return serials, &upgraded, nil
}

func sumDenominations(notes []BanknoteV2) int64 {
	total := int64(0)
	for _, n := range notes {
		total += n.DenominationNg
	}
	return total
}

// --- Leaderboard ---

// LeaderboardEntry represents a holder's position on the banknote value leaderboard.
type LeaderboardEntry struct {
	Holder     string `json:"holder"`
	TotalValue int64  `json:"total_value_ng"`
	NoteCount  int    `json:"note_count"`
}

// GetLeaderboard returns top N holders by total banknote value.
func GetLeaderboard(banknotes []BanknoteV2, n int) []LeaderboardEntry {
	holderValues := map[string]*LeaderboardEntry{}
	for _, b := range banknotes {
		if b.Status != "active" {
			continue
		}
		entry, ok := holderValues[b.Holder]
		if !ok {
			entry = &LeaderboardEntry{Holder: b.Holder}
			holderValues[b.Holder] = entry
		}
		entry.TotalValue += b.DenominationNg * int64(b.Multiplier)
		entry.NoteCount++
	}

	list := make([]LeaderboardEntry, 0, len(holderValues))
	for _, v := range holderValues {
		list = append(list, *v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].TotalValue > list[j].TotalValue
	})
	if n > 0 && n < len(list) {
		list = list[:n]
	}
	return list
}

// --- Auto-Reinvest ---

// AutoReinvestManager handles automatic ng-to-banknote reinvestment for holders.
type AutoReinvestManager struct {
	mu      sync.Mutex
	Enabled map[string]bool `json:"enabled"`
}

// NewAutoReinvestManager creates an auto-reinvest manager with an empty enabled map.
func NewAutoReinvestManager() *AutoReinvestManager {
	return &AutoReinvestManager{Enabled: make(map[string]bool)}
}

// IsEnabled returns whether auto-reinvest is active for the given pubkey.
func (ar *AutoReinvestManager) IsEnabled(pubkey string) bool {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	return ar.Enabled[pubkey]
}

// SetEnabled toggles auto-reinvest for a pubkey.
func (ar *AutoReinvestManager) SetEnabled(pubkey string, enabled bool) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.Enabled[pubkey] = enabled
}

// Reinvest spends ng balance to mint banknotes.
func (ar *AutoReinvestManager) Reinvest(dataDir, pubkey string, amountNg int64) (int, error) {
	if !ar.IsEnabled(pubkey) {
		return 0, fmt.Errorf("auto-reinvest not enabled for %s", pubkey)
	}
	if amountNg < NGPerTLR {
		return 0, fmt.Errorf("minimum reinvest is 1 TLR")
	}

	ledger := LoadLedger(dataDir)
	if ledger.Balance(pubkey) < amountNg {
		return 0, fmt.Errorf("insufficient balance")
	}

	ledger.Mint(pubkey, -amountNg)
	ledger.Save(dataDir)

	now := fmt.Sprintf("%d", time.Now().Unix())
	banknotes, _ := LoadBanknotesV2(dataDir)

	count := int(amountNg / NGPerTLR)
	if count > 10 {
		count = 10
	}
	var newNotes []BanknoteV2
	for i := 0; i < count; i++ {
		newNotes = append(newNotes, BanknoteV2{
			Serial:         fmt.Sprintf("MB-COMMON-REINVEST-%s-%d", pubkey[:6], i),
			DenominationNg: NGPerTLR,
			Rarity:         "common",
			Multiplier:     1,
			Holder:         pubkey,
			FrozenNg:       NGPerTLR,
			Status:         "active",
			MintedAt:       now,
		})
	}

	banknotes = append(banknotes, newNotes...)
	SaveBanknotesV2(dataDir, banknotes)
	return len(newNotes), nil
}
