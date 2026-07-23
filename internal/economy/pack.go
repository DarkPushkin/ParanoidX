// Package economy — Pack Manager: создание и открытие бустер-паков банкнот.
// Бустер содержит 5 банкнот с гарантией минимум 1 Rare+.
// Приоритет наполнения: pre-mit registry (доступные серии) → динамическая генерация на лету.
// PackID формируется по шаблону BP-YYYYMMDD-XXXXX, тип: "booster" или "genesis".
// Паки хранятся в пер-пользовательском инвентаре (inventory-<pubkey>.json).
// При открытии пака банкноты переводятся в статус "active" и закрепляются за владельцем.
package economy

import (
	"fmt"
	"math/rand"
	"time"
)

type PackManager struct{}

// NewPackManager создаёт новый PackManager.
func NewPackManager() *PackManager {
	return &PackManager{}
}

// CreateBoosterPack создаёт бустер-пак из 5 банкнот. Сначала пытается использовать доступные
// pre-mint записи, при их отсутствии генерирует банкноты динамически. Гарантирует минимум
// 1 Rare+ банкноту в пака. Сохраняет банкноты в реестр и пак — в инвентарь владельца.
func (pm *PackManager) CreateBoosterPack(dataDir, ownerPubkey string, preMint []PreMintEntry) (*Pack, error) {
	// try to use pre-mint first
	available := filterAvailable(preMint)

	var banknotes []BanknoteV2
	var serials []string
	var totalFace int64

	if len(available) >= 5 {
		shuffled := make([]PreMintEntry, len(available))
		copy(shuffled, available)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		for i := 0; i < 5; i++ {
			sel := shuffled[i]
			serials = append(serials, sel.Serial)
			totalFace += sel.DenominationNg
			banknotes = append(banknotes, BanknoteV2{
				Serial:         sel.Serial,
				DenominationNg: sel.DenominationNg,
				Rarity:         sel.Rarity,
				Multiplier:     rarityMultiplier[sel.Rarity],
				Holder:         "",
				Status:         "pre_mint",
				MintedAt:       time.Now().UTC().Format(time.RFC3339),
			})
		}

		// guarantee at least 1 rare+
		hasRarePlus := false
		for _, b := range banknotes {
			if b.Rarity != "common" {
				hasRarePlus = true
				break
			}
		}
		if !hasRarePlus {
			rarities := []string{"rare", "epic", "legendary"}
			r := rarities[rand.Intn(len(rarities))]
			for i := range banknotes {
				if banknotes[i].Rarity == "common" {
					repl := pickPreMintByRarity(available, r)
					if repl != nil {
						banknotes[i] = BanknoteV2{
							Serial:         repl.Serial,
							DenominationNg: repl.DenominationNg,
							Rarity:         repl.Rarity,
							Multiplier:     rarityMultiplier[repl.Rarity],
							Holder:         "",
							Status:         "pre_mint",
							MintedAt:       time.Now().UTC().Format(time.RFC3339),
						}
						serials[i] = repl.Serial
						totalFace = totalFace - 0 + repl.DenominationNg
					}
					break
				}
			}
		}

		pmgr := NewPreMintManager()
		pmgr.MarkUsed(dataDir, serials)
	} else {
		// generate on-the-fly
		rarityRoll := func() string {
			r := rand.Intn(100)
			switch {
			case r < 60:
				return "common"
			case r < 85:
				return "rare"
			case r < 95:
				return "epic"
			default:
				return "legendary"
			}
		}

		denoms := []int64{10_000_000_000, 50_000_000_000, 100_000_000_000, 250_000_000_000, 500_000_000_000, 1_000_000_000_000}
		mult := map[string]int{"common": 1, "rare": 2, "epic": 3, "legendary": 4}

		hasRarePlus := false
		for i := 0; i < 5; i++ {
			rarity := rarityRoll()
			if rarity != "common" {
				hasRarePlus = true
			}
			if i == 4 && !hasRarePlus {
				rarity = "rare"
			}
			denom := denoms[rand.Intn(len(denoms))]
			serial := fmt.Sprintf("DYNAMIC-%s-%s-%d", time.Now().Format("20060102"), rarity, rand.Intn(999999))
			banknotes = append(banknotes, BanknoteV2{
				Serial:         serial,
				DenominationNg: denom,
				Rarity:         rarity,
				Multiplier:     mult[rarity],
				Holder:         "",
				Status:         "pre_mint",
				MintedAt:       time.Now().UTC().Format(time.RFC3339),
			})
			serials = append(serials, serial)
			totalFace += denom
		}
	}

	existing, _ := LoadBanknotesV2(dataDir)
	existing = append(existing, banknotes...)
	SaveBanknotesV2(dataDir, existing)

	pack := Pack{
		PackID:    fmt.Sprintf("BP-%s-%d", time.Now().Format("20060102"), rand.Intn(99999)),
		Sealed:    true,
		Banknotes: serials,
		PriceNg:   totalFace,
		Owner:     ownerPubkey,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		PackType:  "booster",
	}

	inv := LoadInventory(dataDir, ownerPubkey)
	inv = append(inv, pack)
	SaveInventory(dataDir, ownerPubkey, inv)
	return &pack, nil
}

// OpenPack открывает запечатанный пак: переводит все банкноты из статуса "pre_mint"
// в "active", назначает владельца, помечает пак как распечатанный.
func (pm *PackManager) OpenPack(dataDir, ownerPubkey, packID string) (*Pack, []BanknoteV2, error) {
	inv := LoadInventory(dataDir, ownerPubkey)
	var pack *Pack
	idx := -1
	for i := range inv {
		if inv[i].PackID == packID && inv[i].Sealed {
			pack = &inv[i]
			idx = i
			break
		}
	}
	if pack == nil {
		return nil, nil, fmt.Errorf("sealed pack not found")
	}

	banknotes, err := LoadBanknotesV2(dataDir)
	if err != nil {
		return nil, nil, fmt.Errorf("load banknotes: %w", err)
	}

	var opened []BanknoteV2
	for _, serial := range pack.Banknotes {
		for j := range banknotes {
			if banknotes[j].Serial == serial && (banknotes[j].Status == "pre_mint" || banknotes[j].Status == "available") {
				banknotes[j].Holder = ownerPubkey
				banknotes[j].Status = "active"
				banknotes[j].FrozenNg = banknotes[j].DenominationNg
				opened = append(opened, banknotes[j])
				break
			}
		}
	}

	if len(opened) == 0 {
		return nil, nil, fmt.Errorf("no banknotes could be opened")
	}

	SaveBanknotesV2(dataDir, banknotes)

	pack.Sealed = false
	inv[idx] = *pack
	SaveInventory(dataDir, ownerPubkey, inv)

	pmgr := NewPreMintManager()
	pmgr.MarkUsed(dataDir, pack.Banknotes)
	return pack, opened, nil
}

// GetUserPacks возвращает все паки пользователя (как запечатанные, так и открытые) из инвентаря.
func (pm *PackManager) GetUserPacks(dataDir, pubkey string) []Pack {
	return LoadInventory(dataDir, pubkey)
}

func filterAvailable(preMint []PreMintEntry) []PreMintEntry {
	var avail []PreMintEntry
	for _, p := range preMint {
		if p.Status == "available" {
			avail = append(avail, p)
		}
	}
	return avail
}

func pickPreMintByRarity(avail []PreMintEntry, rarity string) *PreMintEntry {
	var candidates []PreMintEntry
	for _, p := range avail {
		if p.Rarity == rarity {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	return &candidates[rand.Intn(len(candidates))]
}

// GenerateSerial генерирует серийный номер банкноты в формате MB-<RARITY>-<YYYYMMDD>-<6digit>.
func GenerateSerial(rarity string) string {
	return fmt.Sprintf("MB-%s-%s-%06d", rarity, time.Now().Format("20060102"), rand.Intn(999999))
}
